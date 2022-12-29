const { Axml2xml } = require('axml2xml')

const { data } = require('././data.js')

const buffer = require('buffer')

console.log(Axml2xml.convert)
console.log(buffer)

const [arg0, inputFile ] = process.argv;
console.log(inputFile)

const bufferImpls = {
    nodejs: Buffer,
    npm: buffer.Buffer
}
console.log(bufferImpls)

const originalBuffer = new Buffer(data[0]);
const npmBuffer = new buffer.Buffer(data[0]);

const bufferProxy = new Proxy(originalBuffer, {
    get(target, propKey, receiver) {
        if (propKey === 'length') {
            return target.length
        }
        const item = Reflect.get(target, propKey, receiver);
        if (typeof item === 'function') {
            return function (...args) {
                const npmItem = Reflect.get(npmBuffer, propKey, receiver)
                const npmResult = npmItem.apply(npmBuffer, args)
                const result = item.apply(target, args)
                console.log(propKey, result, npmResult, args)
                return result
            }
        } else {
            return item
        }
    }
})

console.log(bufferProxy)
console.log(bufferProxy.length)
Axml2xml.convert(bufferProxy)
