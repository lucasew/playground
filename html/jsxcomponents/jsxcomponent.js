// Took inspiration from https://dev.to/devalnor/running-jsx-in-your-browser-without-babel-1agc

(function () {
    const elements = document.getElementsByTagName("JSXComponent")
    if (elements.length == 0) {
        console.warn("No JSXComponent found")
        return
    }
    console.log(elements)
    if (!window.React) {
        console.error("React is not defined. Suggested import: https://unpkg.com/react@16/umd/react.production.min.js")
        return
    }
    if (!window.render) {
        console.error("htm is not defined. Suggested import: https://unpkg.com/htm@2.2.1")
        return
    }
    if (!window.ReactDOM) {
        console.error("ReactDOM is not defined. Suggested import: https://unpkg.com/react-dom@16/umd/react-dom.production.min.js")
    }
    const { render } = ReactDOM
    const { createElement, useState, useEffect } = React
    const html = htm.bind(createElement)
    for (let i = 0; i < elements.length; i++) {
        function Component() {
            return eval(element.innerHTML)
        }
        render(html`<${Component}>`, element)
    }
})()
