class RecorderProcessor extends AudioWorkletProcessor {
  constructor() {
    super();
    this.batch = new Float32Array(800);
    this.off = 0;
  }
  process(inputs) {
    const channel = inputs[0]?.[0];
    if (!channel) return true;
    for (let i = 0; i < channel.length; i++) {
      this.batch[this.off++] = channel[i];
      if (this.off >= this.batch.length) {
        this.port.postMessage(this.batch.slice(0));
        this.off = 0;
      }
    }
    return true;
  }
}

registerProcessor("recorder-processor", RecorderProcessor);