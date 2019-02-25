local iterutils = require 'iterutils'
local test = require 'u-test'

test.iterutils_new = function()
    local iter = iterutils.new({1,2,3,4,5})
    test.equal(iter.next(), 1)
    test.equal(iter.next(), 2)
    test.equal(iter.next(), 3)
    test.equal(iter.next(), 4)
    test.equal(iter.next(), 5)
end

test.iterutils_map_double = function()
    local iter = iterutils.new({1,2,3,4,5})
    :map(function (x) return x*2 end)
    test.equal(iter.next(), 2)
    test.equal(iter.next(), 4)
    test.equal(iter.next(), 6)
    test.equal(iter.next(), 8)
    test.equal(iter.next(), 10)
end

test.iterutils_map_pow2 = function()
    local iter = iterutils.new({1,2,3,4,5})
    :map(function (x) return x^2 end)
    test.equal(iter.next(), 1)
    test.equal(iter.next(), 4)
    test.equal(iter.next(), 9)
    test.equal(iter.next(), 16)
    test.equal(iter.next(), 25)
end

test.iterutils_map_div2 = function()
    local iter = iterutils.new({1,2,3,4,5})
    :map(function (x) return x/2 end)
    test.equal(iter.next(), 0.5)
    test.equal(iter.next(), 1)
    test.equal(iter.next(), 1.5)
    test.equal(iter.next(), 2)
    test.equal(iter.next(), 2.5)
end

test.iterutils_filter_even = function()
    local iter = iterutils.new({1,2,3,4,5})
    :filter(function (x) return x%2 == 0 end)
    test.equal(iter.next(), 2)
    test.equal(iter.next(), 4)
end

test.iterutils_reduce_sum = function()
    local res = iterutils.new({1,2,3,4,5})
    :reduce(function (x, y) return x + y end)
    test.equal(res, 15)
end

