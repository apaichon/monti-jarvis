class PlayerProcessor extends AudioWorkletProcessor {
  constructor() {
    super();
    this.queue = [];
    this.cur = null;
    this.curOff = 0;
    this.port.onmessage = (event) => {
      const data = event.data;
      if (data === "flush") {
        this.queue = [];
        this.cur = null;
        this.curOff = 0;
        return;
      }
      if (data instanceof Float32Array) this.queue.push(data);
    };
  }
  process(_inputs, outputs) {
    const out = outputs[0][0];
    if (!out) return true;
    for (let i = 0; i < out.length; i++) {
      if (!this.cur || this.curOff >= this.cur.length) {
        this.cur = this.queue.shift() || null;
        this.curOff = 0;
      }
      out[i] = this.cur ? this.cur[this.curOff++] : 0;
    }
    for (let ch = 1; ch < outputs[0].length; ch++) outputs[0][ch].set(out);
    return true;
  }
}

registerProcessor("player-processor", PlayerProcessor);