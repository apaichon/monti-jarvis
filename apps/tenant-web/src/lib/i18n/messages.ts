export type Lang = 'en' | 'th' | 'ja';

export type Messages = {
  lang_label: string;
  lang_en: string;
  lang_th: string;
  lang_ja: string;
  action_save: string;
  action_cancel: string;
  action_logout: string;
  status_loading: string;
  status_loading_session: string;
  status_all_systems: string;
  brand_console: string;
  workspace: string;
  nav_aria: string;
  nav_group_operations: string;
  nav_overview: string;
  nav_call_center: string;
  nav_monitoring: string;
  nav_tickets: string;
  nav_satisfaction: string;
  nav_preview: string;
  nav_live: string;
  nav_group_knowledge: string;
  nav_knowledge: string;
  nav_gaps: string;
  nav_records: string;
  nav_group_commerce: string;
  nav_billing: string;
  nav_documents: string;
  nav_tax: string;
  nav_group_channels: string;
  nav_avatars: string;
  nav_embed: string;
  nav_theme: string;
  nav_ai_config: string;
  nav_group_directory: string;
  nav_customers: string;
  nav_tiers: string;
  nav_group_growth: string;
  nav_referrals: string;
  nav_group_settings: string;
  nav_settings: string;
  plan_current: string;
  plan_loading: string;
  plan_none: string;
  plan_loading_allowance: string;
  plan_usage_unavailable: string;
  plan_usage_pct: string;
  account_admin: string;
  account_sign_out: string;
  topbar_search: string;
  topbar_notifications: string;
  topbar_profile: string;
  app_version: string;
  // settings page extras
  settings_display_language: string;
  settings_display_language_help: string;
  settings_ai_reply_language: string;
  settings_workspace_locale: string;
  settings_i18n_note: string;
};

const en: Messages = {
  lang_label: 'Language',
  lang_en: 'English',
  lang_th: 'ไทย',
  lang_ja: '日本語',
  action_save: 'Save',
  action_cancel: 'Cancel',
  action_logout: 'Sign out',
  status_loading: 'Loading…',
  status_loading_session: 'Loading session…',
  status_all_systems: 'All systems operational',
  brand_console: 'TENANT CONSOLE',
  workspace: 'Workspace',
  nav_aria: 'Tenant navigation',
  nav_group_operations: 'Operations',
  nav_overview: 'Overview',
  nav_call_center: 'Call center',
  nav_monitoring: 'Monitoring',
  nav_tickets: 'Tickets',
  nav_satisfaction: 'Satisfaction',
  nav_preview: 'Preview',
  nav_live: 'LIVE',
  nav_group_knowledge: 'Knowledge',
  nav_knowledge: 'Knowledge',
  nav_gaps: 'Gaps',
  nav_records: 'Records',
  nav_group_commerce: 'Commerce',
  nav_billing: 'Billing',
  nav_documents: 'Documents',
  nav_tax: 'Tax',
  nav_group_channels: 'Channels',
  nav_avatars: 'Avatars',
  nav_embed: 'Embed',
  nav_theme: 'Theme',
  nav_ai_config: 'AI config',
  nav_group_directory: 'Directory',
  nav_customers: 'Customers',
  nav_tiers: 'Tiers',
  nav_group_growth: 'Growth',
  nav_referrals: 'Referrals',
  nav_group_settings: 'Settings',
  nav_settings: 'Settings',
  plan_current: 'CURRENT PLAN',
  plan_loading: 'Loading…',
  plan_none: 'No active plan',
  plan_loading_allowance: 'Loading allowance…',
  plan_usage_unavailable: 'Usage unavailable',
  plan_usage_pct: '% highest quota usage',
  account_admin: 'Admin',
  account_sign_out: 'Sign out',
  topbar_search: 'Search',
  topbar_notifications: 'Notifications',
  topbar_profile: 'Profile',
  app_version: 'App',
  settings_display_language: 'Display language',
  settings_display_language_help:
    'Controls portal UI labels only (this browser). Does not change AI reply language.',
  settings_ai_reply_language: 'AI reply language',
  settings_workspace_locale: 'Workspace locale',
  settings_i18n_note: 'Display language is separate from AI reply language (Sprint 16).'
};

const th: Messages = {
  ...en,
  lang_label: 'ภาษา',
  action_save: 'บันทึก',
  action_cancel: 'ยกเลิก',
  action_logout: 'ออกจากระบบ',
  status_loading: 'กำลังโหลด…',
  status_loading_session: 'กำลังโหลดเซสชัน…',
  status_all_systems: 'ระบบทำงานปกติ',
  brand_console: 'คอนโซลผู้เช่า',
  workspace: 'พื้นที่ทำงาน',
  nav_aria: 'เมนูผู้เช่า',
  nav_group_operations: 'ปฏิบัติการ',
  nav_overview: 'ภาพรวม',
  nav_call_center: 'คอลเซ็นเตอร์',
  nav_monitoring: 'มอนิเตอร์',
  nav_tickets: 'ตั๋ว',
  nav_satisfaction: 'ความพึงพอใจ',
  nav_preview: 'พรีวิว',
  nav_live: 'สด',
  nav_group_knowledge: 'ความรู้',
  nav_knowledge: 'คลังความรู้',
  nav_gaps: 'ช่องว่าง',
  nav_records: 'บันทึก',
  nav_group_commerce: 'การค้า',
  nav_billing: 'บิลลิ่ง',
  nav_documents: 'เอกสาร',
  nav_tax: 'ภาษี',
  nav_group_channels: 'ช่องทาง',
  nav_avatars: 'อวาตาร์',
  nav_embed: 'ฝังเว็บ',
  nav_theme: 'ธีม',
  nav_ai_config: 'ตั้งค่า AI',
  nav_group_directory: 'ไดเรกทอรี',
  nav_customers: 'ลูกค้า',
  nav_tiers: 'ระดับ',
  nav_group_growth: 'การเติบโต',
  nav_referrals: 'แนะนำ',
  nav_group_settings: 'ตั้งค่า',
  nav_settings: 'ตั้งค่า',
  plan_current: 'แพ็กเกจปัจจุบัน',
  plan_loading: 'กำลังโหลด…',
  plan_none: 'ยังไม่มีแพ็กเกจ',
  plan_loading_allowance: 'กำลังโหลดโควตา…',
  plan_usage_unavailable: 'ไม่พบข้อมูลการใช้งาน',
  plan_usage_pct: '% การใช้โควตาสูงสุด',
  account_admin: 'แอดมิน',
  account_sign_out: 'ออกจากระบบ',
  topbar_search: 'ค้นหา',
  topbar_notifications: 'การแจ้งเตือน',
  topbar_profile: 'โปรไฟล์',
  app_version: 'แอป',
  settings_display_language: 'ภาษาที่แสดงบนหน้าจอ',
  settings_display_language_help:
    'มีผลเฉพาะป้ายกำกับ UI ในเบราว์เซอร์นี้ ไม่เปลี่ยนภาษาที่ AI ตอบ',
  settings_ai_reply_language: 'ภาษาที่ AI ตอบ',
  settings_workspace_locale: 'ภาษาพื้นที่ทำงาน',
  settings_i18n_note: 'ภาษา UI แยกจากภาษาที่ AI ตอบ (Sprint 16)'
};

const ja: Messages = {
  ...en,
  lang_label: '言語',
  action_save: '保存',
  action_cancel: 'キャンセル',
  action_logout: 'サインアウト',
  status_loading: '読み込み中…',
  status_loading_session: 'セッションを読み込み中…',
  status_all_systems: 'すべてのシステムは正常です',
  brand_console: 'テナントコンソール',
  workspace: 'ワークスペース',
  nav_aria: 'テナントナビゲーション',
  nav_group_operations: 'オペレーション',
  nav_overview: '概要',
  nav_call_center: 'コールセンター',
  nav_monitoring: 'モニタリング',
  nav_tickets: 'チケット',
  nav_satisfaction: '満足度',
  nav_preview: 'プレビュー',
  nav_live: 'LIVE',
  nav_group_knowledge: 'ナレッジ',
  nav_knowledge: 'ナレッジ',
  nav_gaps: 'ギャップ',
  nav_records: '記録',
  nav_group_commerce: 'コマース',
  nav_billing: '請求',
  nav_documents: '書類',
  nav_tax: '税',
  nav_group_channels: 'チャネル',
  nav_avatars: 'アバター',
  nav_embed: '埋め込み',
  nav_theme: 'テーマ',
  nav_ai_config: 'AI 設定',
  nav_group_directory: 'ディレクトリ',
  nav_customers: '顧客',
  nav_tiers: 'ティア',
  nav_group_growth: 'グロース',
  nav_referrals: '紹介',
  nav_group_settings: '設定',
  nav_settings: '設定',
  plan_current: '現在のプラン',
  plan_loading: '読み込み中…',
  plan_none: '有効なプランなし',
  plan_loading_allowance: '枠を読み込み中…',
  plan_usage_unavailable: '利用状況を取得できません',
  plan_usage_pct: '% 最大クォータ使用率',
  account_admin: '管理者',
  account_sign_out: 'サインアウト',
  topbar_search: '検索',
  topbar_notifications: '通知',
  topbar_profile: 'プロフィール',
  app_version: 'アプリ',
  settings_display_language: '表示言語',
  settings_display_language_help:
    'このブラウザの UI ラベルのみ変更します。AI の返信言語は変わりません。',
  settings_ai_reply_language: 'AI 返信言語',
  settings_workspace_locale: 'ワークスペース言語',
  settings_i18n_note: '表示言語は AI 返信言語と別です（Sprint 16）。'
};

export const messages: Record<Lang, Messages> = { en, th, ja };
