const init = function() {
    const injectElement = document.createElement('div');
    injectElement.id = 'injectElement';
    injectElement.innerHTML = 'Injected Element';
    document.body.appendChild(injectElement);
}
init();