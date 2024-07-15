let lastQrData;
let lastQr;

try {
	let ws = new WebSocket("/code");
	ws.onmessage = (event) => {
		/** @type {string} */
		let text = event.data;
		if (lastQrData === text) return; // no change
		console.log("QR code contents:", text);
		try {
			if (lastQr) {
				console.log("updating qr code");
				lastQr.clear();
				lastQr.makeCode(text);
			} else
				lastQr = new QRCode(document.getElementById("qrcode"), {
					text,
					width: 512,
					height: 512,
					correctLevel: QRCode.CorrectLevel.M,
				});
			lastQrData = text;
		} catch (err) {
			alert("couldn't make QR code: " + err.message);
			console.error(err);
		}
	};
} catch (err) {
	alert("couldn't connect to server: " + err.message);
	console.error(err);
}
