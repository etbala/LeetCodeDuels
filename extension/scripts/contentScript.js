


if(typeof init === 'undefined'){
  const init = function(){
    const injectElement = document.createElement('div');
    injectElement.className = 'rustyZone-element';
    injectElement.innerHTML = 'Hello From the Rusty Zone Element';
    document.body.appendChild(injectElement);

    const hostEle = document.createElement('div');
    hostEle.className = 'rustyZone-element-host';
    hostEle.innerHTML = 'Hello From the Rusty Zone Element';
    document.body.appendChild(hostEle);

    //Using Shadow Root
    var host = document.querySelector('.rustyZone-element-host');
    var root = host.attachShadow({mode: 'open'}); // Create a Shadow Root
    var div = document.createElement('div');
    div.className = 'div root-class';
    div.innerHTML='<style>.div{border:3px solid blue;margin:10px;padding:10px;width:200px;}</style>'
    +'Hello From the Rusty Zone Shadow Root Element';
    root.appendChild(div);
  }
  init();
}
























    /*
    const hostEle = document.createElement('div');
    hostEle.className = 'rustyZone-element-host';
    hostEle.innerHTML = 'Hello From the Rusty Zone Element';
    document.body.appendChild(hostEle);

    //Using Shadow Root
    var host = document.querySelector('.rustyZone-element-host');
    var root = host.attachShadow({mode: 'open'}); // Create a Shadow Root
    var div = document.createElement('div');
    div.className = 'div root-class';
    div.innerHTML='<style>.div{border:3px solid blue;margin:10px;padding:10px;width:200px;}</style>'
    +'Hello From the Rusty Zone Shadow Root Element';
    root.appendChild(div);
    */
