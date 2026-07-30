export type Lang = 'en' | 'th' | 'ja';

export type Messages = {
  lang_label: string;
  lang_en: string;
  lang_th: string;
  lang_ja: string;
  action_save: string;
  action_cancel: string;
  action_search: string;
  action_logout: string;
  action_send: string;
  status_loading: string;
  status_empty: string;
  status_error: string;
  status_live: string;
  status_ok: string;
  status_online: string;
  status_limited: string;
  status_offline: string;
  status_checking: string;

  // Directory
  picker_aria: string;
  picker_title: string;
  picker_sub: string;
  picker_search: string;
  picker_search_ph: string;
  picker_loading: string;
  picker_empty_title: string;
  picker_empty_body: string;
  picker_call: string;
  picker_badge: string;
  picker_showing: string;
  picker_foot_secure: string;
  picker_foot_ai: string;
  picker_load_error: string;

  // Desk chrome
  desk_tagline: string;
  desk_selected_tenant: string;
  desk_change_tenant: string;
  desk_active: string;
  desk_brand_line: string;
  desk_agent: string;
  desk_customer: string;
  desk_end: string;
  desk_end_call: string;
  desk_start_call: string;
  desk_connecting: string;
  desk_ready: string;
  desk_live_now: string;
  desk_ai_avatar_live: string;
  desk_audio_settings: string;
  desk_audio_help: string;
  desk_mic: string;
  desk_speaker: string;
  desk_default_speaker: string;
  desk_speaker_note: string;
  desk_refresh_devices: string;
  desk_test_audio: string;
  desk_test_help: string;
  desk_start_test: string;
  desk_stop_test: string;
  desk_ai_avatar_call: string;
  desk_choose_avatar: string;
  desk_sign_in_avatars: string;
  desk_my_tenants: string;
  desk_manage_tenants: string;
  desk_add_tenant: string;
  desk_account: string;
  desk_sign_out: string;
  desk_sign_in: string;
  desk_email: string;
  desk_name: string;
  desk_otp: string;
  desk_send_otp: string;
  desk_verify: string;
  desk_footer_secure: string;
  desk_composer_ph: string;
  desk_topic_general: string;
  desk_topic_billing: string;
  desk_topic_technical: string;
  desk_welcome: string;
  desk_voice_hint: string;
  desk_new_session: string;
  desk_expand: string;
  desk_collapse: string;
};

const en: Messages = {
  lang_label: 'Language',
  lang_en: 'English',
  lang_th: 'ไทย',
  lang_ja: '日本語',
  action_save: 'Save',
  action_cancel: 'Cancel',
  action_search: 'Search',
  action_logout: 'Sign out',
  action_send: 'Send',
  status_loading: 'Loading…',
  status_empty: 'Nothing here yet',
  status_error: 'Something went wrong',
  status_live: 'Live',
  status_ok: 'OK',
  status_online: 'Online',
  status_limited: 'Limited',
  status_offline: 'Offline',
  status_checking: 'Checking…',

  picker_aria: 'Choose a brand to call',
  picker_title: 'Choose who to call',
  picker_sub: 'Select a brand to start chat or voice',
  picker_search: 'Search brands',
  picker_search_ph: 'Search brands…',
  picker_loading: 'Loading brands…',
  picker_empty_title: 'No brands available',
  picker_empty_body: 'No public brands are listed yet.',
  picker_call: 'Call',
  picker_badge: 'AI · Text & Voice',
  picker_showing: 'Showing',
  picker_foot_secure: 'Secure · Private · Enterprise-grade',
  picker_foot_ai: 'AI-powered conversations',
  picker_load_error: 'Failed to load brands',

  desk_tagline: 'Your AI assistant. Always here to help.',
  desk_selected_tenant: 'Selected tenant',
  desk_change_tenant: 'Change tenant',
  desk_active: 'Active',
  desk_brand_line: 'Brand',
  desk_agent: 'Agent',
  desk_customer: 'Customer',
  desk_end: 'End',
  desk_end_call: 'End call',
  desk_start_call: 'Start call',
  desk_connecting: 'Connecting…',
  desk_ready: 'Ready to call',
  desk_live_now: 'Live now',
  desk_ai_avatar_live: 'AI AVATAR LIVE',
  desk_audio_settings: 'AUDIO SETTINGS',
  desk_audio_help: 'Configure your microphone and speaker',
  desk_mic: 'Microphone',
  desk_speaker: 'Speaker',
  desk_default_speaker: 'Default speaker',
  desk_speaker_note: 'Speaker selection needs Chrome/Edge (setSinkId). Mic selection still works.',
  desk_refresh_devices: 'Refresh devices',
  desk_test_audio: 'Test your audio',
  desk_test_help: 'Make sure your mic and speaker work properly.',
  desk_start_test: 'Start test',
  desk_stop_test: 'Stop test',
  desk_ai_avatar_call: 'AI AVATAR CALL',
  desk_choose_avatar: 'Choose an AI avatar to start a conversation',
  desk_sign_in_avatars: 'Sign in to show available AI avatars.',
  desk_my_tenants: 'MY TENANTS',
  desk_manage_tenants: 'Manage and switch between your tenants',
  desk_add_tenant: 'Add tenant',
  desk_account: 'Account',
  desk_sign_out: 'Sign out',
  desk_sign_in: 'Sign in',
  desk_email: 'Email',
  desk_name: 'Display name',
  desk_otp: 'OTP code',
  desk_send_otp: 'Send OTP',
  desk_verify: 'Verify',
  desk_footer_secure: 'Secure connection',
  desk_composer_ph: 'Type a message…',
  desk_topic_general: 'General',
  desk_topic_billing: 'Billing',
  desk_topic_technical: 'Technical',
  desk_welcome:
    'Welcome to Monti Inbound Call Center. Choose an AI agent on the left, then type a question or start a voice call.',
  desk_voice_hint: 'Select an agent, then start an inbound voice call.',
  desk_new_session: 'New call session',
  desk_expand: 'Expand',
  desk_collapse: 'Collapse'
};

const th: Messages = {
  ...en,
  lang_label: 'ภาษา',
  action_save: 'บันทึก',
  action_cancel: 'ยกเลิก',
  action_search: 'ค้นหา',
  action_logout: 'ออกจากระบบ',
  action_send: 'ส่ง',
  status_loading: 'กำลังโหลด…',
  status_empty: 'ยังไม่มีข้อมูล',
  status_error: 'เกิดข้อผิดพลาด',
  status_live: 'สด',
  status_ok: 'ปกติ',
  status_online: 'ออนไลน์',
  status_limited: 'จำกัด',
  status_offline: 'ออฟไลน์',
  status_checking: 'กำลังตรวจสอบ…',

  picker_aria: 'เลือกแบรนด์ที่ต้องการโทร',
  picker_title: 'เลือกปลายทางที่ต้องการติดต่อ',
  picker_sub: 'เลือกแบรนด์เพื่อเริ่มแชทหรือสายเสียง',
  picker_search: 'ค้นหาแบรนด์',
  picker_search_ph: 'ค้นหาแบรนด์…',
  picker_loading: 'กำลังโหลดแบรนด์…',
  picker_empty_title: 'ยังไม่มีแบรนด์',
  picker_empty_body: 'ยังไม่มีแบรนด์ที่เปิดให้บริการ',
  picker_call: 'โทร',
  picker_badge: 'AI · ข้อความและเสียง',
  picker_showing: 'แสดง',
  picker_foot_secure: 'ปลอดภัย · เป็นส่วนตัว · ระดับองค์กร',
  picker_foot_ai: 'สนทนาด้วย AI',
  picker_load_error: 'โหลดแบรนด์ไม่สำเร็จ',

  desk_tagline: 'ผู้ช่วย AI ของคุณ พร้อมช่วยเหลือเสมอ',
  desk_selected_tenant: 'แบรนด์ที่เลือก',
  desk_change_tenant: 'เปลี่ยนแบรนด์',
  desk_active: 'ใช้งานอยู่',
  desk_brand_line: 'แบรนด์',
  desk_agent: 'เอเจนต์',
  desk_customer: 'ลูกค้า',
  desk_end: 'จบ',
  desk_end_call: 'วางสาย',
  desk_start_call: 'เริ่มสาย',
  desk_connecting: 'กำลังเชื่อมต่อ…',
  desk_ready: 'พร้อมโทร',
  desk_live_now: 'กำลังสนทนา',
  desk_ai_avatar_live: 'อวาตาร์สด',
  desk_audio_settings: 'ตั้งค่าเสียง',
  desk_audio_help: 'เลือกไมโครโฟนและลำโพง',
  desk_mic: 'ไมโครโฟน',
  desk_speaker: 'ลำโพง',
  desk_default_speaker: 'ลำโพงเริ่มต้น',
  desk_speaker_note: 'การเลือกลำโพงต้องใช้ Chrome/Edge (setSinkId) ยังเลือกไมค์ได้ตามปกติ',
  desk_refresh_devices: 'รีเฟรชอุปกรณ์',
  desk_test_audio: 'ทดสอบเสียง',
  desk_test_help: 'ตรวจสอบว่าไมค์และลำโพงใช้งานได้',
  desk_start_test: 'เริ่มทดสอบ',
  desk_stop_test: 'หยุดทดสอบ',
  desk_ai_avatar_call: 'โทรด้วยอวาตาร์ AI',
  desk_choose_avatar: 'เลือกอวาตาร์ AI เพื่อเริ่มสนทนา',
  desk_sign_in_avatars: 'เข้าสู่ระบบเพื่อดูอวาตาร์ AI',
  desk_my_tenants: 'แบรนด์ของฉัน',
  desk_manage_tenants: 'จัดการและสลับแบรนด์',
  desk_add_tenant: 'เพิ่มแบรนด์',
  desk_account: 'บัญชี',
  desk_sign_out: 'ออกจากระบบ',
  desk_sign_in: 'เข้าสู่ระบบ',
  desk_email: 'อีเมล',
  desk_name: 'ชื่อที่แสดง',
  desk_otp: 'รหัส OTP',
  desk_send_otp: 'ส่ง OTP',
  desk_verify: 'ยืนยัน',
  desk_footer_secure: 'การเชื่อมต่อปลอดภัย',
  desk_composer_ph: 'พิมพ์ข้อความ…',
  desk_topic_general: 'ทั่วไป',
  desk_topic_billing: 'การเงิน',
  desk_topic_technical: 'เทคนิค',
  desk_welcome:
    'ยินดีต้อนรับสู่ศูนย์บริการโทรเข้า Monti เลือกเอเจนต์ทางซ้าย แล้วพิมพ์คำถามหรือเริ่มสายเสียง',
  desk_voice_hint: 'เลือกเอเจนต์ แล้วเริ่มสายเสียงขาเข้า',
  desk_new_session: 'เซสชันการโทรใหม่',
  desk_expand: 'ขยาย',
  desk_collapse: 'ย่อ'
};

const ja: Messages = {
  ...en,
  lang_label: '言語',
  action_save: '保存',
  action_cancel: 'キャンセル',
  action_search: '検索',
  action_logout: 'サインアウト',
  action_send: '送信',
  status_loading: '読み込み中…',
  status_empty: 'まだありません',
  status_error: 'エラーが発生しました',
  status_live: 'ライブ',
  status_ok: '正常',
  status_online: 'オンライン',
  status_limited: '制限あり',
  status_offline: 'オフライン',
  status_checking: '確認中…',

  picker_aria: '通話するブランドを選択',
  picker_title: '通話先を選ぶ',
  picker_sub: 'チャットまたは音声を開始するブランドを選択',
  picker_search: 'ブランドを検索',
  picker_search_ph: 'ブランドを検索…',
  picker_loading: 'ブランドを読み込み中…',
  picker_empty_title: 'ブランドがありません',
  picker_empty_body: '公開ブランドはまだ登録されていません。',
  picker_call: '通話',
  picker_badge: 'AI · テキスト＆音声',
  picker_showing: '表示中',
  picker_foot_secure: 'セキュア · プライベート · エンタープライズ',
  picker_foot_ai: 'AI による会話',
  picker_load_error: 'ブランドの読み込みに失敗しました',

  desk_tagline: 'あなたの AI アシスタント。いつでもサポートします。',
  desk_selected_tenant: '選択中のテナント',
  desk_change_tenant: 'テナントを変更',
  desk_active: '有効',
  desk_brand_line: 'ブランド',
  desk_agent: 'エージェント',
  desk_customer: '顧客',
  desk_end: '終了',
  desk_end_call: '通話終了',
  desk_start_call: '通話開始',
  desk_connecting: '接続中…',
  desk_ready: '通話準備完了',
  desk_live_now: '通話中',
  desk_ai_avatar_live: 'AI アバター ライブ',
  desk_audio_settings: '音声設定',
  desk_audio_help: 'マイクとスピーカーを設定',
  desk_mic: 'マイク',
  desk_speaker: 'スピーカー',
  desk_default_speaker: 'デフォルトのスピーカー',
  desk_speaker_note:
    'スピーカー選択には Chrome/Edge（setSinkId）が必要です。マイク選択は利用できます。',
  desk_refresh_devices: 'デバイスを更新',
  desk_test_audio: '音声テスト',
  desk_test_help: 'マイクとスピーカーが動作するか確認します。',
  desk_start_test: 'テスト開始',
  desk_stop_test: 'テスト停止',
  desk_ai_avatar_call: 'AI アバター通話',
  desk_choose_avatar: '会話を始める AI アバターを選択',
  desk_sign_in_avatars: 'サインインして AI アバターを表示',
  desk_my_tenants: 'マイテナント',
  desk_manage_tenants: 'テナントの管理と切り替え',
  desk_add_tenant: 'テナントを追加',
  desk_account: 'アカウント',
  desk_sign_out: 'サインアウト',
  desk_sign_in: 'サインイン',
  desk_email: 'メール',
  desk_name: '表示名',
  desk_otp: 'OTP コード',
  desk_send_otp: 'OTP を送信',
  desk_verify: '確認',
  desk_footer_secure: 'セキュア接続',
  desk_composer_ph: 'メッセージを入力…',
  desk_topic_general: '一般',
  desk_topic_billing: '請求',
  desk_topic_technical: '技術',
  desk_welcome:
    'Monti インバウンドコールセンターへようこそ。左の AI エージェントを選び、質問を入力するか音声通話を開始してください。',
  desk_voice_hint: 'エージェントを選び、インバウンド音声通話を開始してください。',
  desk_new_session: '新しい通話セッション',
  desk_expand: '展開',
  desk_collapse: '折りたたむ'
};

export const messages: Record<Lang, Messages> = { en, th, ja };
