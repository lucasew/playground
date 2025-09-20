local pandoc = require 'pandoc'
local utils = require 'pandoc.utils'
local stringify = utils.stringify

local gdoc = {}
local meta = {}

function Str(el) do
    -- print("chegou aqui = ", el.text)
    local text = el.text
    local rep = meta[text]
    if rep ~= nil then
        text = rep
    end
    return pandoc.Str(text)
end
end

function Pandoc(doc) do
    gdoc = doc
    for k, v in pairs(doc.meta) do
        meta["%" .. k .. "%"] = stringify(v)
        -- print(k, " = ", stringify(v))
    end
    return doc
end
end

function CodeBlock(el) do
    local lang = el.attr.classes[1]
    local text = el.text
    if lang == "evalme" then
        fn = load(text)
        return fn()
    else
        -- return pandoc.Para(text .. ":" .. lang)
    end
end
end

return {
    { Pandoc = Pandoc },
    { Str = Str },
    { CodeBlock = CodeBlock },
}
