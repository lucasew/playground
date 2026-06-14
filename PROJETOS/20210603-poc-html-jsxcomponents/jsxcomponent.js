// Took inspiration from https://dev.to/devalnor/running-jsx-in-your-browser-without-babel-1agc

(function () {
    window.captureAppError = window.captureAppError || function(err, context) {
        console.warn("[App Error]", err, context || '');
        if (typeof Sentry !== 'undefined') {
            Sentry.captureException(err, { extra: context });
        }
    };

    const elements = document.getElementsByTagName("JSXComponent")
    if (elements.length == 0) {
        console.warn("No JSXComponent found")
        return
    }
    console.log(elements)
    if (!window.React) {
        window.captureAppError(new Error("React is not defined. Suggested import: https://unpkg.com/react@16/umd/react.production.min.js"))
        return
    }
    if (!window.render) {
        window.captureAppError(new Error("htm is not defined. Suggested import: https://unpkg.com/htm@2.2.1"))
        return
    }
    if (!window.ReactDOM) {
        window.captureAppError(new Error("ReactDOM is not defined. Suggested import: https://unpkg.com/react-dom@16/umd/react-dom.production.min.js"))
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
