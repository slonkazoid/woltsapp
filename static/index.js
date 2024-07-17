/**
 * @typedef {object} Message
 * @property {number} id
 * @property {string} body
 */

const header = document.getElementById("header");
const qrcode = document.getElementById("qrcode");
const done = document.getElementById("done");

let lastQrData;
let lastQr;
let loggedIn = false;

function processQr(text) {
	if (lastQrData === text) return; // no change
	console.log("QR code contents:", text);
	try {
		if (lastQr) {
			console.log("updating qr code");
			lastQr.clear();
			lastQr.makeCode(text);
		} else
			lastQr = new QRCode(document.getElementById("qrcode"), {
				text: text,
				width: 512,
				height: 512,
				correctLevel: QRCode.CorrectLevel.M,
			});
		lastQrData = text;
	} catch (err) {
		alert("couldn't make QR code: " + err.message);
		console.error(err);
	}
}

function main() {
	let ws = new WebSocket("/code");
	ws.onmessage = (event) => {
		/** @type {Message} */
		let message;
		try {
			message = JSON.parse(event.data);
		} catch (err) {
			alert("error: malformed message; check console for more details");
			console.error(err);
		}

		let { id, body } = message;
		console.log("got message id %d, %d bytes", id, body.length);

		switch (id) {
			case 0:
				processQr(body);
				break;
			case 1:
				console.log("logged in");
				header.style.opacity = "50%";
				qrcode.style.opacity = "50%";
				qrcode.style.cursor = "not-allowed";
				done.style.display = "block";
				loggedIn = true;

				ws.close();
				break;
		}
	};
	ws.onerror = (event) => {
		console.error("ws error:", event);
	};
	ws.onclose = (event) => {
		console.log("ws close:", event);
		if (!loggedIn) {
			console.warn("websocket closed before logged in, reconnecting in 5 seconds...");
			setTimeout(main, 5000);
		}
	};
}

main();
