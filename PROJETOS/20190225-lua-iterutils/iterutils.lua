require 'table'

local iterutils = {}
iterutils._name = "iterutils"

-- Adiciona map filter reduce na galera reaproveitando código
function iterutils._wrap_iter(data)
    function data:map(f)
        local res = {}
        function res.next()
            local v = self.next()
            if v == nil then
                return nil
            end
            return f(v)
        end
        return iterutils._wrap_iter(res)
    end
    function data:filter(f)
        local res = {}
        res.next = function()
            local r = self:next()
            while r ~= nil do
                if f(r) then
                    return r
                end
                r = self:next()
            end
        end
        return iterutils._wrap_iter(res)
    end
    function data:reduce(f)
        local initial = self:next()
        local v = self:next()
        while v ~= nil do
            initial = f(initial, v)
            v = self:next()
        end
        return initial
    end
    return data
end

-- Apenas implementa next
function iterutils.new(data)
    local iter = {}
    function iter:next()
        return table.remove(data, 1)
    end
    return iterutils._wrap_iter(iter)
end

-- Example
--[[
local itr = iterutils.new({1,2,3,4})
:map(function(x) return x^2 end)
:filter(function(x) return x%2 == 0 end)

local v = itr:next()
while v do
    print(v)
    v = itr:next()
end
]]--

return iterutils
