'use strict';

class OBSWebSocket extends EventTarget {
	constructor(url, password) {
		super()

		this.promises = {}
		this.id = 0

		this.ws = new WebSocket(url)
		this.ws.addEventListener("open", e => this.authenticate(password)
			.then(e => this.dispatchEvent(new CustomEvent("open", {}))))
		
		this.ws.addEventListener("message", e => {
			const m = JSON.parse(e.data)
			const id = m["message-id"]
			if (id) {
				const p = this.promises[id]
				if (p) {
					delete this.promises[id]
					switch (m.status) {
					case "error":
						p.reject(new Error(m.error))
					case "ok":
						p.resolve(m)
					}
				}
				return
			}
			if ('update-type' in m) {
				this.dispatchEvent(new CustomEvent(m["update-type"], {detail: m}))
			}
		})
		this.ws.addEventListener("close", e => this.dispatchEvent(new CustomEvent("close", {})))
	}

	send(requestType, msg) {
		const id = String(this.id++)
		msg["request-type"] = requestType
		msg["message-id"] = id
		return new Promise((resolve, reject) => {
			this.promises[id] = {resolve, reject}
			this.ws.send(JSON.stringify(msg))
		})
	}

	async getCurrentScene() { 
		return this.send("GetCurrentScene", {})
	}

	setMute(source, mute) {
		this.send("SetMute", {source, mute})
	}

	toggleMute(source) {
		this.send("ToggleMute", {source})
	}

	takeSourceScreenshot(sourceName, embedPictureFormat = "png", width = 480, height = 270) {
		return this.send("TakeSourceScreenshot", {sourceName, embedPictureFormat, width, height})
	}

	async authenticate(password) {
		const {authRequired, salt, challenge} = await this.send("GetAuthRequired", {})
		if (!authRequired) return true

		const e = new TextEncoder()
		const h1 = await crypto.subtle.digest('SHA-256', e.encode(password.concat(salt)))
		const b64 = btoa(String.fromCharCode(...new Uint8Array(h1)))
		const h2 = await crypto.subtle.digest('SHA-256', e.encode(b64.concat(challenge)))
		return this.send("Authenticate", {auth: btoa(String.fromCharCode(...new Uint8Array(h2)))})
	}

	close() {
		this.ws.close()
	}
}
