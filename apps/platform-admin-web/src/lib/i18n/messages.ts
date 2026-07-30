export type Lang = 'en' | 'th' | 'ja';

export type Messages = {
  lang_label: string;
  lang_en: string;
  lang_th: string;
  lang_ja: string;
  action_logout: string;
  brand_admin: string;
  nav_aria: string;
  nav_overview: string;
  nav_packages: string;
  nav_tenants: string;
  nav_avatars: string;
  nav_billing: string;
  nav_quotes: string;
  nav_leads: string;
  nav_audit: string;
  nav_monitoring: string;
  nav_call_center: string;
  nav_payment: string;
  nav_profile: string;
  system_health: string;
  system_all_ok: string;
  account_admin: string;
  account_sign_out: string;
  topbar_platform: string;
  topbar_admin: string;
  topbar_search: string;
  topbar_notifications: string;
  role_super: string;
  admin_version: string;
};

const en: Messages = {
  lang_label: 'Language',
  lang_en: 'English',
  lang_th: 'ไทย',
  lang_ja: '日本語',
  action_logout: 'Sign out',
  brand_admin: 'PLATFORM ADMIN',
  nav_aria: 'Platform navigation',
  nav_overview: 'Overview',
  nav_packages: 'Packages',
  nav_tenants: 'Tenants',
  nav_avatars: 'Avatars',
  nav_billing: 'Billing',
  nav_quotes: 'Quote requests',
  nav_leads: 'Leads',
  nav_audit: 'Audit log',
  nav_monitoring: 'Monitoring',
  nav_call_center: 'Call center',
  nav_payment: 'Payment',
  nav_profile: 'Profile',
  system_health: 'System health',
  system_all_ok: 'All systems operational',
  account_admin: 'Admin',
  account_sign_out: 'Sign out',
  topbar_platform: 'Monti Platform',
  topbar_admin: 'Administration',
  topbar_search: 'Search',
  topbar_notifications: 'Notifications',
  role_super: 'SUPER ADMIN',
  admin_version: 'Monti Admin'
};

const th: Messages = {
  ...en,
  lang_label: 'ภาษา',
  action_logout: 'ออกจากระบบ',
  brand_admin: 'ผู้ดูแลแพลตฟอร์ม',
  nav_aria: 'เมนูแพลตฟอร์ม',
  nav_overview: 'ภาพรวม',
  nav_packages: 'แพ็กเกจ',
  nav_tenants: 'ผู้เช่า',
  nav_avatars: 'อวาตาร์',
  nav_billing: 'บิลลิ่ง',
  nav_quotes: 'คำขอใบเสนอราคา',
  nav_leads: 'ลีด',
  nav_audit: 'บันทึกตรวจสอบ',
  nav_monitoring: 'มอนิเตอร์',
  nav_call_center: 'คอลเซ็นเตอร์',
  nav_payment: 'ชำระเงิน',
  nav_profile: 'โปรไฟล์',
  system_health: 'สุขภาพระบบ',
  system_all_ok: 'ระบบทำงานปกติ',
  account_admin: 'แอดมิน',
  account_sign_out: 'ออกจากระบบ',
  topbar_platform: 'Monti แพลตฟอร์ม',
  topbar_admin: 'การดูแลระบบ',
  topbar_search: 'ค้นหา',
  topbar_notifications: 'การแจ้งเตือน',
  role_super: 'ซูเปอร์แอดมิน',
  admin_version: 'Monti Admin'
};

const ja: Messages = {
  ...en,
  lang_label: '言語',
  action_logout: 'サインアウト',
  brand_admin: 'プラットフォーム管理',
  nav_aria: 'プラットフォームナビゲーション',
  nav_overview: '概要',
  nav_packages: 'パッケージ',
  nav_tenants: 'テナント',
  nav_avatars: 'アバター',
  nav_billing: '請求',
  nav_quotes: '見積リクエスト',
  nav_leads: 'リード',
  nav_audit: '監査ログ',
  nav_monitoring: 'モニタリング',
  nav_call_center: 'コールセンター',
  nav_payment: '支払い',
  nav_profile: 'プロフィール',
  system_health: 'システム状態',
  system_all_ok: 'すべてのシステムは正常です',
  account_admin: '管理者',
  account_sign_out: 'サインアウト',
  topbar_platform: 'Monti プラットフォーム',
  topbar_admin: '管理',
  topbar_search: '検索',
  topbar_notifications: '通知',
  role_super: 'スーパー管理者',
  admin_version: 'Monti Admin'
};

export const messages: Record<Lang, Messages> = { en, th, ja };
