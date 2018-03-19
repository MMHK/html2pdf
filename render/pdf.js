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
    width: '1240px', 
    height: '1754px', 
    margin: '0px'
}
/* page.clipRect = { 
    top: 0, 
    left: 0, 
    width: 1240, 
    height: 1754 
};
 */
page.zoomFactor = 1;

page.open(address, function (status) {
    if (status !== 'success') {
        console.log('Unable to load the address!');
        phantom.exit(1);
    } else {
        window.setTimeout(function () {
            page.render(output, {format: 'pdf', quality: '10'});
            phantom.exit();
        }, 200);
    }
});