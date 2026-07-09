// Response-tone classification for avatar expressions. Mirrors the tones in
// internal/workforce ExpressionTones; returns '' when no tone stands out.
export type Tone = 'hello' | 'happy' | 'sorry' | 'cheer' | 'goodbye' | '';

const rules: Array<{ tone: Tone; pattern: RegExp }> = [
  {
    tone: 'goodbye',
    pattern:
      /\bgood\s?bye\b|\bbye\b|see you|take care|have a (great|good|nice) (day|one|week)|ลาก่อน|แล้วเจอกัน|แล้วพบกัน|โชคดีนะ|ขอบคุณที่ใช้บริการ/i
  },
  {
    tone: 'sorry',
    pattern: /\bsorry\b|\bapolog|\bregret\b|unfortunately|ขอโทษ|ขออภัย|ต้องขออภัย|เสียใจ/i
  },
  {
    tone: 'cheer',
    pattern:
      /congrat|awesome|fantastic|excellent|well done|great news|wonderful|🎉|🥳|ยินดีด้วย|เยี่ยมมาก|สุดยอด|เก่งมาก/i
  },
  {
    tone: 'hello',
    pattern:
      /^(hi|hello|hey)\b|\bwelcome\b|สวัสดี|ยินดีต้อนรับ|thank you for calling|how (can|may) i help/i
  },
  {
    tone: 'happy',
    pattern: /\bglad\b|happy to|my pleasure|ด้วยความยินดี|ยินดีช่วย|ยินดีที่ได้|😊|:\)/i
  }
];

export function classifyTone(text: string): Tone {
  const t = (text || '').trim();
  if (!t) return '';
  for (const rule of rules) {
    if (rule.pattern.test(t)) return rule.tone;
  }
  return '';
}
