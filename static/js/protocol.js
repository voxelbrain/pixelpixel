var ws = new WebSocket('ws://'+window.location.host+'/ws');

var i = 0;
var type_created = i++;
var type_change  = i++;
var type_failure = i++;

ws.onclose = function() {
	console.log('WebSocket closed');
};
ws.onmessage = function(ev) {
	event = JSON.parse(ev.data);
	switch(event.type) {
		case type_created:
		var pixel = document.createElement('pixel-pixel');
		pixel.id = event.pixel;
		pixel.key = event.pixel;
		document.body.appendChild(pixel);
		break;
		case type_change:
		var pixel = document.getElementById(event.pixel);
		if(pixel)
			pixel.update();
		break;
		case type_failure:
			console.log('ERROR');
		break
	}
};
