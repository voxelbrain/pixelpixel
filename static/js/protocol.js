window.addEventListener('WebComponentsReady', function() {
	var pw = document.getElementsByTagName('pixel-pixelwall')[0];
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
				pw.pixels += ' '+event.pixel;
			break;
			case type_change:
				pw.update(event.pixel);
			break;
			case type_failure:
				pw.markCrashed(event.pixel);
			break
		}
	};
});
