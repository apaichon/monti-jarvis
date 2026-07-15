export const CUSTOMER_END_COUNTDOWN_SECONDS = 5;

/** Detect a caller confirmation that there is nothing else to discuss. */
export function customerConfirmedEnd(text: string): boolean {
  const normalized = (text || '')
    .normalize('NFC')
    .toLowerCase()
    .replace(/[\s!?.,…:;"'“”‘’()\[\]{}-]+/gu, '');

  return [
    /ไม่มี(?:แล้ว|อีก)?(?:นะ)?(?:ครับ|ค่ะ|คะ|จ้ะ|จ้า)?ขอบคุณ(?:มาก)?(?:ครับ|ค่ะ|นะครับ|นะคะ)?/u,
    /ไม่มีแล้ว(?:นะ)?(?:ครับ|ค่ะ|คะ|จ้ะ|จ้า)?/u,
    /ไม่มีอะไรแล้ว(?:นะ)?(?:ครับ|ค่ะ|คะ)?/u,
    /หมดคำถามแล้ว(?:นะ)?(?:ครับ|ค่ะ|คะ)?/u,
    /nomore(?:questions?)?/,
    /that'sall(?:thankyou)?/,
    /nomorequestions(?:thankyou)?/,
    /nothingelse(?:thankyou)?/,
    /that'sit(?:thankyou)?/
  ].some((pattern) => pattern.test(normalized));
}

/** Detect the assistant's farewell after the caller confirmed the end. */
export function assistantConfirmedFarewell(text: string): boolean {
  const normalized = (text || '').toLowerCase();
  return /ขออนุญาต.*(?:วางสาย|ปิดสาย)|(?:วางสาย|ปิดสาย).*ขอบคุณ|goodbye|good\s?bye|call will close/i.test(
    normalized
  );
}
