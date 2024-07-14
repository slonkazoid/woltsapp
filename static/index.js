let lastQrData;
let lastQr;

async function fetchAndRender() {
	try {
		let res = await fetch("/code");
		if (res.status === 200) {
			let text = await res.text();
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
		} else {
			alert(`unexpected ${res.status} response`);
			console.error(res);
		}
	} catch (err) {
		if (err.name === "NetworkError") alert();
		alert("couldn't retrieve code: " + err.message);
		console.error(err);
	}
}

//await fetchAndRender();
//setInterval(fetchAndRender, 1000);

new QRCode(document.getElementById("qrcode"), {
	text: "a".repeat(179),
	width: 512,
	height: 512,
	correctLevel: QRCode.CorrectLevel.M,
});
