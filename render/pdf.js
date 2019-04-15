"use strict";

var page = require('webpage').create(),
    system = require('system'),
    address, 
    output = "pdf";
    
if (system.args.length > 2) {
    address = system.args[1];
    output = system.args[2];
}

page.paperSize = {
    margin: '0px',
    format: "A4"
}
page.zoomFactor = 1;
page.open(address, function (status) {
    if (status !== 'success') {
        console.log('Unable to load the address!');
        phantom.exit(1);
    } else {
        window.setTimeout(function () {
            page.evaluate(function(){
                document.body.style.zoom = 0.48;
            });
            page.render(output, {format: 'pdf', quality: '10'});
            phantom.exit();
        }, 200);
    }
});