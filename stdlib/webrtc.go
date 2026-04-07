package stdlib

// WebRTCRuntime provides peer-to-peer messaging helpers.
const WebRTCRuntime = `
// Quill WebRTC Runtime
// Peer-to-peer messaging that bypasses the server

class PeerConnection {
  constructor(config) {
    this.pc = new RTCPeerConnection(config || {
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }]
    });
    this.channel = null;
    this._onMessage = null;
    this._onConnect = null;
    this._onClose = null;
    this.pc.onicecandidate = (e) => {
      if (e.candidate && this._onIceCandidate) {
        this._onIceCandidate(JSON.stringify(e.candidate));
      }
    };
  }

  createChannel(name) {
    this.channel = this.pc.createDataChannel(name || "quill");
    this._setupChannel(this.channel);
    return this;
  }

  _setupChannel(channel) {
    channel.onopen = () => { if (this._onConnect) this._onConnect(); };
    channel.onclose = () => { if (this._onClose) this._onClose(); };
    channel.onmessage = (e) => { if (this._onMessage) this._onMessage(e.data); };
  }

  onMessage(fn) { this._onMessage = fn; return this; }
  onConnect(fn) { this._onConnect = fn; return this; }
  onClose(fn) { this._onClose = fn; return this; }
  onIceCandidate(fn) { this._onIceCandidate = fn; return this; }

  async createOffer() {
    this.createChannel("quill");
    const offer = await this.pc.createOffer();
    await this.pc.setLocalDescription(offer);
    return JSON.stringify(offer);
  }

  async acceptOffer(offerStr) {
    const offer = JSON.parse(offerStr);
    await this.pc.setRemoteDescription(offer);
    this.pc.ondatachannel = (e) => {
      this.channel = e.channel;
      this._setupChannel(this.channel);
    };
    const answer = await this.pc.createAnswer();
    await this.pc.setLocalDescription(answer);
    return JSON.stringify(answer);
  }

  async acceptAnswer(answerStr) {
    const answer = JSON.parse(answerStr);
    await this.pc.setRemoteDescription(answer);
  }

  async addIceCandidate(candidateStr) {
    const candidate = JSON.parse(candidateStr);
    await this.pc.addIceCandidate(candidate);
  }

  send(data) {
    if (this.channel && this.channel.readyState === "open") {
      this.channel.send(typeof data === "string" ? data : JSON.stringify(data));
    }
  }

  close() {
    if (this.channel) this.channel.close();
    this.pc.close();
  }
}

function createPeer(config) {
  return new PeerConnection(config);
}
`

// GetWebRTCRuntime returns the WebRTC runtime string.
func GetWebRTCRuntime() string {
	return WebRTCRuntime
}
