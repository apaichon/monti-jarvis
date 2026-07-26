export type Lang = 'en' | 'th' | 'ja';

export type Messages = {
  brand_tagline: string;
  nav_product: string;
  nav_solutions: string;
  nav_solutions_industry: string;
  nav_solutions_enterprise: string;
  nav_resources: string;
  nav_pricing: string;
  nav_about: string;
  nav_login: string;
  nav_book_demo: string;
  nav_contact: string;
  nav_demo: string;
  nav_open_menu: string;
  lang_label: string;
  footer_blurb: string;
  footer_product: string;
  footer_overview: string;
  footer_get_started: string;
  footer_live_demo: string;
  footer_register: string;
  footer_contact_sales: string;
  footer_company: string;
  footer_demo_guide: string;
  footer_privacy: string;
  footer_rights: string;
  footer_care: string;

  // Home
  home_title: string;
  home_pill: string;
  home_h1_1: string;
  home_h1_2: string;
  home_h1_3: string;
  home_lede: string;
  home_cta_demo: string;
  home_cta_video: string;
  home_chip_voice: string;
  home_chip_lang: string;
  home_chip_km: string;
  home_chip_secure: string;
  home_ava_role: string;
  home_ava_tone: string;
  home_desk_title: string;
  home_desk_live: string;
  home_desk_general: string;
  home_desk_billing: string;
  home_desk_tech: string;
  home_desk_welcome: string;
  home_desk_you: string;
  home_desk_user_msg: string;
  home_desk_just_now: string;
  home_desk_placeholder: string;
  home_desk_send: string;
  home_stat_60: string;
  home_stat_1000s: string;
  home_stat_247: string;
  home_stat_csat: string;
  home_built_eyebrow: string;
  home_built_h2: string;
  home_cap_voice_t: string;
  home_cap_voice_b: string;
  home_cap_km_t: string;
  home_cap_km_b: string;
  home_cap_omni_t: string;
  home_cap_omni_b: string;
  home_cap_workflow_t: string;
  home_cap_workflow_b: string;
  home_cap_handover_t: string;
  home_cap_handover_b: string;
  home_cap_insights_t: string;
  home_cap_insights_b: string;
  home_use_eyebrow: string;
  home_use_h2: string;
  home_use_1: string;
  home_use_2: string;
  home_use_3: string;
  home_use_4: string;
  home_use_5: string;
  home_use_6: string;
  home_qr_h3: string;
  home_qr_p: string;
  home_qr_cta: string;
  home_trust: string;
  home_trust_note: string;
  home_ready_h2: string;
  home_ready_p: string;
  home_ready_demo: string;
  home_ready_book: string;
  home_ready_register: string;

  // Product
  product_title: string;
  product_h1: string;
  product_lede: string;
  product_cta_demo: string;
  product_cta_register: string;
  product_suite_kicker: string;
  product_nav_overview: string;
  product_nav_voice: string;
  product_nav_omni: string;
  product_nav_km: string;
  product_nav_handover: string;
  product_nav_analytics: string;
  product_nav_security: string;
  product_nav_integrations: string;
  product_km_title: string;
  product_km_sub: string;
  product_km_search: string;
  product_km_1: string;
  product_km_2: string;
  product_km_3: string;
  product_voice_talking: string;
  product_voice_role: string;
  product_voice_end: string;
  product_pitch_h2: string;
  product_pitch_em: string;
  product_pitch_p: string;
  product_ben_easy_t: string;
  product_ben_easy_b: string;
  product_ben_scale_t: string;
  product_ben_scale_b: string;
  product_ben_flex_t: string;
  product_ben_flex_b: string;
  product_ben_secure_t: string;
  product_ben_secure_b: string;
  product_footer_eyebrow: string;
  product_footer_h2: string;
  product_footer_demo: string;
  product_footer_talk: string;

  // Solutions
  solutions_title: string;
  solutions_h1: string;
  solutions_p: string;
  solutions_learn: string;
  solutions_ind_cs_t: string;
  solutions_ind_cs_b: string;
  solutions_ind_fin_t: string;
  solutions_ind_fin_b: string;
  solutions_ind_ecom_t: string;
  solutions_ind_ecom_b: string;
  solutions_ind_health_t: string;
  solutions_ind_health_b: string;
  solutions_ind_travel_t: string;
  solutions_ind_travel_b: string;
  solutions_ind_tel_t: string;
  solutions_ind_tel_b: string;
  solutions_ind_edu_t: string;
  solutions_ind_edu_b: string;
  solutions_ind_gov_t: string;
  solutions_ind_gov_b: string;
  solutions_expert_h2: string;
  solutions_expert_p: string;
  solutions_expert_cta: string;

  // Solutions — Enterprise
  ent_title: string;
  ent_badge: string;
  ent_h1_1: string;
  ent_h1_2: string;
  ent_h1_3: string;
  ent_lede: string;
  ent_cta_sales: string;
  ent_cta_arch: string;
  ent_layer_ai: string;
  ent_layer_ai_sub: string;
  ent_layer_data: string;
  ent_layer_data_sub: string;
  ent_layer_env: string;
  ent_own_monti: string;
  ent_own_you: string;
  ent_secure_link: string;
  ent_trust_title: string;
  ent_trust_1: string;
  ent_trust_2: string;
  ent_trust_3: string;
  ent_trust_4: string;
  ent_how_h2: string;
  ent_how_1_t: string;
  ent_how_1_b: string;
  ent_how_2_t: string;
  ent_how_2_b: string;
  ent_how_3_t: string;
  ent_how_3_b: string;
  ent_how_4_t: string;
  ent_how_4_b: string;
  ent_managed_h2: string;
  ent_managed_foot: string;
  ent_managed_1: string;
  ent_managed_2: string;
  ent_managed_3: string;
  ent_managed_4: string;
  ent_managed_5: string;
  ent_controlled_h2: string;
  ent_controlled_foot: string;
  ent_controlled_1: string;
  ent_controlled_2: string;
  ent_controlled_3: string;
  ent_controlled_4: string;
  ent_controlled_5: string;
  ent_deploy_h2: string;
  ent_deploy_1_t: string;
  ent_deploy_1_b: string;
  ent_deploy_2_t: string;
  ent_deploy_2_b: string;
  ent_deploy_3_t: string;
  ent_deploy_3_b: string;
  ent_ben_h2: string;
  ent_ben_1_t: string;
  ent_ben_1_b: string;
  ent_ben_2_t: string;
  ent_ben_2_b: string;
  ent_ben_3_t: string;
  ent_ben_3_b: string;
  ent_ben_4_t: string;
  ent_ben_4_b: string;
  ent_ben_5_t: string;
  ent_ben_5_b: string;
  ent_ben_6_t: string;
  ent_ben_6_b: string;
  ent_cta_h2: string;
  ent_cta_p: string;
  ent_cta_demo: string;
  ent_cta_architect: string;
  ent_arch_web: string;
  ent_arch_voice: string;
  ent_arch_chat: string;
  ent_arch_orch: string;
  ent_arch_guard: string;
  ent_arch_analytics: string;
  ent_arch_calls: string;
  ent_arch_km: string;
  ent_arch_customer: string;
  ent_arch_storage: string;
  ent_arch_db: string;
  ent_arch_audit: string;

  // Resources (embed SDK + mobile branded call)
  resources_title: string;
  resources_h1: string;
  resources_p: string;
  resources_web_eyebrow: string;
  resources_web_h2: string;
  resources_install: string;
  resources_example: string;
  resources_copy: string;
  resources_copied: string;
  resources_hint: string;
  resources_mobile_eyebrow: string;
  resources_mobile_h2: string;
  resources_mobile_p: string;
  resources_mobile_f1_t: string;
  resources_mobile_f1_b: string;
  resources_mobile_f2_t: string;
  resources_mobile_f2_b: string;
  resources_mobile_f3_t: string;
  resources_mobile_f3_b: string;
  resources_mobile_cap_brands: string;
  resources_mobile_cap_tenant: string;
  resources_mobile_cap_call: string;
  resources_mobile_sdk_h3: string;
  resources_mobile_sdk_p: string;

  // Pricing
  pricing_title: string;
  pricing_h1: string;
  pricing_p: string;
  pricing_monthly: string;
  pricing_annual: string;
  pricing_save: string;
  pricing_loading: string;
  pricing_error: string;
  pricing_error_hint: string;
  pricing_most_popular: string;
  pricing_choose: string;
  pricing_contact_sales: string;
  pricing_catalog_note: string;
  pricing_no_setup_t: string;
  pricing_no_setup_b: string;
  pricing_cancel_t: string;
  pricing_cancel_b: string;
  pricing_secure_t: string;
  pricing_secure_b: string;
  pricing_support_t: string;
  pricing_support_b: string;
  pricing_legal: string;
  pricing_perfect_start: string;
  pricing_growing: string;
  pricing_scaling: string;
  pricing_large: string;
  pricing_custom: string;
  pricing_blurb_starter: string;
  pricing_blurb_growth: string;
  pricing_blurb_pro: string;
  pricing_blurb_ent: string;
  pricing_shared_h2: string;
  pricing_shared_p: string;
  pricing_dedicated_h2: string;
  pricing_dedicated_p: string;
  pricing_request_quote: string;
  pricing_buy_now: string;

  // About
  about_title: string;
  about_h1: string;
  about_p1: string;
  about_p2: string;
  about_story: string;
  about_stat_team: string;
  about_stat_customers: string;
  about_stat_convos: string;
  about_stat_support: string;
  about_values: string;
  about_val_customer_t: string;
  about_val_customer_b: string;
  about_val_innov_t: string;
  about_val_innov_b: string;
  about_val_integ_t: string;
  about_val_integ_b: string;
  about_val_impact_t: string;
  about_val_impact_b: string;
  about_cta_h2: string;
  about_cta_p: string;
  about_cta_btn: string;

  // Contact
  contact_title: string;
  contact_book: string;
  contact_sales: string;
  contact_newsletter: string;
  contact_lede: string;
  contact_thanks: string;
  contact_deduped: string;
  contact_received: string;
  contact_see_demo: string;
  contact_back_home: string;
  contact_type_book: string;
  contact_type_contact: string;
  contact_name: string;
  contact_email: string;
  contact_company: string;
  contact_phone: string;
  contact_usecase: string;
  contact_channel: string;
  contact_channel_email: string;
  contact_channel_phone: string;
  contact_channel_line: string;
  contact_channel_other: string;
  contact_consent_contact: string;
  contact_consent_marketing: string;
  contact_submit: string;
  contact_submitting: string;
  contact_err_email: string;
  contact_err_consent: string;
  contact_err_marketing: string;
  contact_err_generic: string;

  // Demo
  demo_title: string;
  demo_badge: string;
  demo_h1: string;
  demo_lede: string;
  demo_try_h2: string;
  demo_try_1: string;
  demo_try_2: string;
  demo_try_3: string;
  demo_try_4: string;
  demo_after_h2: string;
  demo_after_1: string;
  demo_after_2: string;
  demo_after_3: string;
  demo_after_4: string;
  demo_open_h2: string;
  demo_open_p: string;
  demo_launch: string;
  demo_book: string;
  demo_register: string;
};

export const messages: Record<Lang, Messages> = {
  en: {
    brand_tagline: 'AI CALL CENTER',
    nav_product: 'Product',
    nav_solutions: 'Solutions',
    nav_solutions_industry: 'Every industry',
    nav_solutions_enterprise: 'Enterprise',
    nav_resources: 'Resources',
    nav_pricing: 'Pricing',
    nav_about: 'About',
    nav_login: 'Login',
    nav_book_demo: 'Book a demo',
    nav_contact: 'Contact',
    nav_demo: 'Demo',
    nav_open_menu: 'Open menu',
    lang_label: 'Language',
    footer_blurb: 'AI call-center workforce for modern support teams.',
    footer_product: 'Product',
    footer_overview: 'Overview',
    footer_get_started: 'Get started',
    footer_live_demo: 'Live demo',
    footer_register: 'Register',
    footer_contact_sales: 'Contact sales',
    footer_company: 'Company',
    footer_demo_guide: 'Demo guide',
    footer_privacy: 'Privacy-aware lead capture · no payment on marketing pages',
    footer_rights: 'Monti. All rights reserved.',
    footer_care: 'Built for teams that answer with care.',

    home_title: 'Monti — AI Conversations That Understand',
    home_pill: '✦ AI Voice Agents for Modern Businesses',
    home_h1_1: 'AI Conversations',
    home_h1_2: 'That Understand.',
    home_h1_3: 'Outcomes That Matter.',
    home_lede:
      'Monti is your 24/7 AI call center workforce. Human-like voice agents that understand, respond, and resolve — across every channel and language.',
    home_cta_demo: 'Try Monti Live Demo',
    home_cta_video: 'Watch video',
    home_chip_voice: 'Voice AI Agents',
    home_chip_lang: 'Multi-language Support',
    home_chip_km: 'Knowledge Powered',
    home_chip_secure: 'Secure & Compliant',
    home_ava_role: 'General Support',
    home_ava_tone: 'Warm & Patient',
    home_desk_title: 'Monti Caller Desk',
    home_desk_live: 'Live',
    home_desk_general: 'General',
    home_desk_billing: 'Billing',
    home_desk_tech: 'Technical',
    home_desk_welcome: 'Welcome to Monti Inbound Call Center. How can I help you today?',
    home_desk_you: 'You',
    home_desk_user_msg: 'I need help with my billing.',
    home_desk_just_now: 'Just now',
    home_desk_placeholder: 'Ask your question…',
    home_desk_send: 'Send',
    home_stat_60: 'Reduce call handling time',
    home_stat_1000s: 'Calls handled simultaneously',
    home_stat_247: 'Always on, never miss a call',
    home_stat_csat: 'Happier customers, better experience',
    home_built_eyebrow: 'Built for business',
    home_built_h2: 'Everything you need to deliver great experiences',
    home_cap_voice_t: 'Human-like AI Voice',
    home_cap_voice_b: 'Natural conversations that feel human',
    home_cap_km_t: 'Smart Knowledge',
    home_cap_km_b: 'Answers from your data, policies, and docs',
    home_cap_omni_t: 'Omnichannel',
    home_cap_omni_b: 'Voice, web widget, mobile, and more',
    home_cap_workflow_t: 'Workflow Ready',
    home_cap_workflow_b: 'Integrate with your systems and tools',
    home_cap_handover_t: 'Live Handover',
    home_cap_handover_b: 'Seamless transfer to human agents',
    home_cap_insights_t: 'Insights & Analytics',
    home_cap_insights_b: 'Real-time monitoring and actionable insights',
    home_use_eyebrow: 'Use cases',
    home_use_h2: 'Monti works where you do',
    home_use_1: 'Customer Support',
    home_use_2: 'Billing & Payments',
    home_use_3: 'Technical Support',
    home_use_4: 'Sales & Lead Qualification',
    home_use_5: 'Appointment Booking',
    home_use_6: 'Kiosk & Self-Service',
    home_qr_h3: 'See Monti in action',
    home_qr_p: 'Experience a live AI voice conversation in your language.',
    home_qr_cta: 'Try Live Demo',
    home_trust: 'Trusted by businesses',
    home_trust_note: 'Placeholder brands for layout only — not real endorsements.',
    home_ready_h2: 'Ready when you are',
    home_ready_p:
      'Experience the product live, talk to sales, or open a tenant workspace and pick a package from the catalog.',
    home_ready_demo: 'Try live demo',
    home_ready_book: 'Book a demo',
    home_ready_register: 'Start free registration',

    product_title: 'Product — Monti Product Suite',
    product_h1: 'Monti Product Suite',
    product_lede: 'Everything you need to build, run, and scale AI conversations that deliver real results.',
    product_cta_demo: 'Try a live demo',
    product_cta_register: 'Start free registration',
    product_suite_kicker: 'Product suite',
    product_nav_overview: 'Overview',
    product_nav_voice: 'AI Voice Agents',
    product_nav_omni: 'Omnichannel',
    product_nav_km: 'Knowledge Hub',
    product_nav_handover: 'Live Handover',
    product_nav_analytics: 'Analytics & Insights',
    product_nav_security: 'Security & Compliance',
    product_nav_integrations: 'Integrations',
    product_km_title: 'Knowledge Hub',
    product_km_sub: 'Answers grounded in your business',
    product_km_search: 'Search knowledge…',
    product_km_1: 'Billing FAQ',
    product_km_2: 'Return & Refund Policy',
    product_km_3: 'Technical Guide',
    product_voice_talking: 'Talking with Ava',
    product_voice_role: 'AI Voice Agent',
    product_voice_end: 'End call',
    product_pitch_h2: 'Powerful by design.',
    product_pitch_em: 'Simple to use.',
    product_pitch_p:
      'Monti combines advanced AI with an intuitive experience so your team can focus on what matters most — your customers.',
    product_ben_easy_t: 'Easy to start',
    product_ben_easy_b: 'Get up and running in minutes, not weeks.',
    product_ben_scale_t: 'Built for scale',
    product_ben_scale_b: 'Enterprise-grade platform ready to grow with you.',
    product_ben_flex_t: 'Flexible & open',
    product_ben_flex_b: 'Works the way you do with open integrations.',
    product_ben_secure_t: 'Reliable & secure',
    product_ben_secure_b: 'Built with security, privacy, and compliance at the core.',
    product_footer_eyebrow: 'See the whole picture',
    product_footer_h2: 'One suite. Every conversation.',
    product_footer_demo: 'Explore Monti live',
    product_footer_talk: 'Talk to our team',

    solutions_title: 'Solutions — Monti',
    solutions_h1: 'AI conversations for every industry and team',
    solutions_p: 'Monti adapts to your business needs and delivers better experiences across every touchpoint.',
    solutions_learn: 'Learn more',
    solutions_ind_cs_t: 'Customer Support',
    solutions_ind_cs_b: 'Resolve customer issues faster with AI agents that understand and respond naturally.',
    solutions_ind_fin_t: 'Financial Services',
    solutions_ind_fin_b: 'Secure, compliant, and reliable conversations for banking, insurance, and fintech.',
    solutions_ind_ecom_t: 'E-commerce',
    solutions_ind_ecom_b: 'Improve sales and customer satisfaction with 24/7 AI shopping assistants.',
    solutions_ind_health_t: 'Healthcare',
    solutions_ind_health_b: 'Provide patient support and appointment help with empathy and accuracy.',
    solutions_ind_travel_t: 'Travel & Hospitality',
    solutions_ind_travel_b: 'Answer guest questions, manage bookings, and deliver delightful experiences.',
    solutions_ind_tel_t: 'Telecom',
    solutions_ind_tel_b: 'Reduce churn and improve CX with intelligent self-service and support.',
    solutions_ind_edu_t: 'Education',
    solutions_ind_edu_b: 'Engage students and parents with smart information and support bots.',
    solutions_ind_gov_t: 'Public Sector',
    solutions_ind_gov_b: 'Deliver better citizen services with secure and accessible AI interactions.',
    solutions_expert_h2: 'Not sure where to start?',
    solutions_expert_p: 'Our team can help you find the right solution for your business.',
    solutions_expert_cta: 'Talk to an expert',

    ent_title: 'Enterprise Solutions — Monti',
    ent_badge: 'Solutions for Enterprise',
    ent_h1_1: 'Enterprise AI with',
    ent_h1_2: 'Data Ownership',
    ent_h1_3: 'Built In',
    ent_lede:
      'Keep the AI experience managed by Monti while deploying your call records, knowledge base, and business data on your on-premise or private cloud environment.',
    ent_cta_sales: 'Talk to sales',
    ent_cta_arch: 'See architecture',
    ent_layer_ai: 'AI EXPERIENCE LAYER / AIAAS BY MONTI',
    ent_layer_ai_sub: 'Managed, updated and secured by Monti',
    ent_layer_data: 'CUSTOMER DATA LAYER / YOUR ENVIRONMENT',
    ent_layer_data_sub: 'Deployed, stored and controlled by you',
    ent_layer_env: 'ON-PREMISE / PRIVATE CLOUD / CUSTOMER VPC',
    ent_own_monti: 'Managed by Monti',
    ent_own_you: 'Owned by You',
    ent_secure_link: 'Secure connectivity over encrypted channel',
    ent_trust_title: 'Your data stays yours.',
    ent_trust_1: 'Own your call records',
    ent_trust_2: 'Keep knowledge base in your environment',
    ent_trust_3: 'Deploy on-premise or private cloud',
    ent_trust_4: 'PDPA / enterprise-ready architecture',
    ent_how_h2: 'HOW IT WORKS',
    ent_how_1_t: '1. Managed AI Frontend',
    ent_how_1_b: 'Monti delivers the AI experience — voice agents, chat UI, orchestration, guardrails, and analytics as a managed service.',
    ent_how_2_t: '2. Secure Connector / API Gateway',
    ent_how_2_b: 'All interactions flow through a secure gateway with strong authentication, encryption, and access controls.',
    ent_how_3_t: '3. On-Premise Data Layer',
    ent_how_3_b: 'Your data layer stays in your environment — call records, knowledge base, databases, and audit logs.',
    ent_how_4_t: '4. Flexible Deployment & Setup',
    ent_how_4_b: 'Deploy on-premise, private cloud, or hybrid — with setup options that fit your IT and compliance requirements.',
    ent_managed_h2: 'MANAGED BY MONTI',
    ent_managed_foot: 'Monti manages the AI experience so you can focus on your business.',
    ent_managed_1: 'AI Agents',
    ent_managed_2: 'Voice Interface',
    ent_managed_3: 'Orchestration',
    ent_managed_4: 'Continuous Updates',
    ent_managed_5: 'Support & Success',
    ent_controlled_h2: 'CONTROLLED BY YOU',
    ent_controlled_foot: 'You own your data, policies and access — end to end.',
    ent_controlled_1: 'Call Records',
    ent_controlled_2: 'KM Documents',
    ent_controlled_3: 'Storage & Databases',
    ent_controlled_4: 'Retention Policy',
    ent_controlled_5: 'Access Control',
    ent_deploy_h2: 'DEPLOYMENT OPTIONS',
    ent_deploy_1_t: 'On-Premise',
    ent_deploy_1_b: 'Deploy the data layer within your on-premise data center with full control and isolation.',
    ent_deploy_2_t: 'Private Cloud',
    ent_deploy_2_b: 'Run in your private cloud environment with dedicated resources and enterprise-grade security.',
    ent_deploy_3_t: 'Hybrid',
    ent_deploy_3_b: 'Combine on-premise security for sensitive data with cloud flexibility for other workloads.',
    ent_ben_h2: 'ENTERPRISE BENEFITS',
    ent_ben_1_t: 'Data Ownership',
    ent_ben_1_b: 'You own and control all your data',
    ent_ben_2_t: 'Security & Compliance',
    ent_ben_2_b: 'Built for enterprise standards and PDPA',
    ent_ben_3_t: 'Faster Rollout',
    ent_ben_3_b: 'Pre-built AI experience with quick deployment',
    ent_ben_4_t: 'Scalable Architecture',
    ent_ben_4_b: 'Built to grow with your organization',
    ent_ben_5_t: 'Integration-Ready',
    ent_ben_5_b: 'Connects with your existing systems',
    ent_ben_6_t: 'Auditability',
    ent_ben_6_b: 'Full visibility with audit logs and access tracking',
    ent_cta_h2: 'Let’s design the right enterprise deployment for your organization.',
    ent_cta_p: 'Our team will work with your IT and security teams to create a deployment model that meets your requirements and protects what matters most.',
    ent_cta_demo: 'Book enterprise demo',
    ent_cta_architect: 'Talk to architect',
    ent_arch_web: 'Web Widget',
    ent_arch_voice: 'Voice Agent',
    ent_arch_chat: 'Chat UI',
    ent_arch_orch: 'Orchestration',
    ent_arch_guard: 'Guardrails',
    ent_arch_analytics: 'Analytics Dashboard',
    ent_arch_calls: 'Call Records',
    ent_arch_km: 'Knowledge Base (KM)',
    ent_arch_customer: 'Customer Data',
    ent_arch_storage: 'Storage',
    ent_arch_db: 'Database',
    ent_arch_audit: 'Audit Logs',

    resources_title: 'Embed & Mobile SDKs — Monti',
    resources_h1: 'Embed & Mobile SDKs',
    resources_p:
      'Ship Monti on your website with framework embeds, and on mobile with your own brand so customers can search tenants and place direct AI calls.',
    resources_web_eyebrow: 'Web embed',
    resources_web_h2: 'Website SDKs by technology',
    resources_install: 'Install',
    resources_example: 'Example embed code',
    resources_copy: 'Copy code',
    resources_copied: 'Copied',
    resources_hint:
      'Requires a Monti server origin (api-base / apiBase) and an active embed key with your host origin allowlisted.',
    resources_mobile_eyebrow: 'Mobile · your brand',
    resources_mobile_h2: 'Custom brand mobile app for customer direct calls',
    resources_mobile_p:
      'Beyond web embed, Monti provides a mobile path where customers open a branded experience: browse public tenant brands, pick an AI agent, and start chat or voice calls directly — with your company name, avatars, and locale (EN/TH/JA).',
    resources_mobile_f1_t: 'Brand directory',
    resources_mobile_f1_b: 'Customers search and open listed tenant brands from a public mobile directory.',
    resources_mobile_f2_t: 'Your brand surface',
    resources_mobile_f2_b: 'Tenant name, agents, language, and quota appear under your brand before the call starts.',
    resources_mobile_f3_t: 'Direct AI call',
    resources_mobile_f3_b: 'Start call or chat straight from mobile — branded avatar, live voice, and transcript.',
    resources_mobile_cap_brands: 'Find brands / tenants',
    resources_mobile_cap_tenant: 'Tenant brand + AI agents',
    resources_mobile_cap_call: 'Live branded call',
    resources_mobile_sdk_h3: 'Mobile SDK example',
    resources_mobile_sdk_p:
      'TypeScript core for iOS, Android, React Native, and Flutter hosts. Auth, avatars, quota, and call lifecycle stay server-owned.',

    pricing_title: 'Pricing — Monti',
    pricing_h1: 'Simple, transparent pricing',
    pricing_p: 'Choose the plan that fits your business. Upgrade or downgrade anytime.',
    pricing_monthly: 'Monthly',
    pricing_annual: 'Annual',
    pricing_save: 'Save 20%',
    pricing_loading: 'Loading packages…',
    pricing_error: 'Package catalog is temporarily unavailable.',
    pricing_error_hint: 'Showing reference plans — confirm live prices after registration or contact sales.',
    pricing_most_popular: 'Most popular',
    pricing_choose: 'Choose',
    pricing_contact_sales: 'Contact sales',
    pricing_catalog_note: 'Live prices from package catalog · entitlement starts only after checkout',
    pricing_no_setup_t: 'No setup fees',
    pricing_no_setup_b: 'Get started in minutes.',
    pricing_cancel_t: 'Cancel anytime',
    pricing_cancel_b: 'No long-term contracts.',
    pricing_secure_t: 'Secure & compliant',
    pricing_secure_b: 'Enterprise-grade security.',
    pricing_support_t: 'Dedicated support',
    pricing_support_b: "We're here to help you.",
    pricing_legal:
      'Call minutes and usage are subject to our Fair Usage Policy. Prices exclude applicable taxes. Public pricing never grants quota — payment remains in authenticated tenant billing.',
    pricing_perfect_start: 'Perfect for getting started',
    pricing_growing: 'For growing businesses',
    pricing_scaling: 'For scaling teams',
    pricing_large: 'For large organizations',
    pricing_custom: 'Custom',
    pricing_blurb_starter: 'Perfect for getting started',
    pricing_blurb_growth: 'For growing businesses',
    pricing_blurb_pro: 'For scaling teams',
    pricing_blurb_ent: 'For large organizations',
    pricing_shared_h2: 'Shared Cloud — startups & SME',
    pricing_shared_p:
      'Multi-tenant platform. Buy online with the payment gateway. Bring your own AI API keys; platform voice minutes are unlimited.',
    pricing_dedicated_h2: 'Dedicated VM — request a quote',
    pricing_dedicated_p:
      'Isolated infrastructure for larger orgs. Request a quote so we can confirm available server capacity before provisioning.',
    pricing_request_quote: 'Request quote',
    pricing_buy_now: 'Buy with payment',

    about_title: 'About us — Monti',
    about_h1: 'We Build AI that feels human conversation',
    about_p1: 'Natural conversations. Smarter resolutions. Stronger relationships.',
    about_p2:
      'Our team brings together deep expertise in AI, voice technology, and customer experience to build solutions that truly make a difference.',
    about_story: 'Watch the conversation — open live demo',
    about_stat_team: 'Team members',
    about_stat_customers: 'Happy customers',
    about_stat_convos: 'Conversations powered monthly',
    about_stat_support: 'Support & success',
    about_values: 'Why Monti',
    about_val_customer_t: 'Natural Voice AI',
    about_val_customer_b: 'Human-like conversations that build trust.',
    about_val_innov_t: 'Efficient & Smart',
    about_val_innov_b: 'Resolve more with context, not scripts.',
    about_val_integ_t: 'Human + AI',
    about_val_integ_b: 'Seamless handoffs. Stronger together.',
    about_val_impact_t: 'Secure & Compliant',
    about_val_impact_b: 'Enterprise-grade security you can rely on.',
    about_cta_h2: "Let's create better conversations together.",
    about_cta_p: "We'd love to hear about your challenges and explore how Monti can help your business grow.",
    about_cta_btn: 'Book a demo',

    contact_title: 'Contact — Monti',
    contact_book: 'Book a demo',
    contact_sales: 'Contact sales',
    contact_newsletter: 'Stay in the loop',
    contact_lede:
      'Tell us about your inbound support goals. We only follow up when you consent. No payment is processed on this page.',
    contact_thanks: 'Thanks — sales will follow up.',
    contact_deduped: 'We already have a recent request from this email for the same intent.',
    contact_received: 'Your request was received. A teammate will reach out on your preferred channel when available.',
    contact_see_demo: 'See live demo options',
    contact_back_home: 'Back to home',
    contact_type_book: 'Book a demo',
    contact_type_contact: 'Contact',
    contact_name: 'Full name',
    contact_email: 'Work email',
    contact_company: 'Company',
    contact_phone: 'Phone',
    contact_usecase: 'Use case',
    contact_channel: 'Preferred channel',
    contact_channel_email: 'Email',
    contact_channel_phone: 'Phone',
    contact_channel_line: 'LINE',
    contact_channel_other: 'Other',
    contact_consent_contact: 'I agree to be contacted about Monti products and demos.',
    contact_consent_marketing: 'I want product updates and marketing emails (optional).',
    contact_submit: 'Submit',
    contact_submitting: 'Submitting…',
    contact_err_email: 'Email is required.',
    contact_err_consent: 'Contact consent is required.',
    contact_err_marketing: 'Marketing consent is required for newsletter signup.',
    contact_err_generic: 'Unable to submit. Please try again shortly.',

    demo_title: 'Live demo — Monti',
    demo_badge: 'LIVE DEMO',
    demo_h1: 'Experience Monti AI avatars without an account',
    demo_lede:
      'The live demo is the existing no-auth customer portal. Pick an AI avatar agent and ask questions by text or voice. Safe attribution parameters from this product site are preserved when you continue.',
    demo_try_h2: 'What you can try',
    demo_try_1: 'Select a branded AI avatar agent',
    demo_try_2: 'Ask product-style questions over text',
    demo_try_3: 'Optional voice interaction where enabled',
    demo_try_4: 'See how inbound AI coverage feels for visitors',
    demo_after_h2: 'After the demo',
    demo_after_1: 'Book a guided walkthrough with sales',
    demo_after_2: 'Register a tenant workspace',
    demo_after_3: 'Choose a package from the live catalog',
    demo_after_4: 'Checkout remains in authenticated billing',
    demo_open_h2: 'Open the live demo',
    demo_open_p: 'You will leave the marketing site for the customer demo surface at the site root.',
    demo_launch: 'Launch live demo',
    demo_book: 'Book a guided demo',
    demo_register: 'Start registration'
  },

  th: {
    brand_tagline: 'ศูนย์บริการ AI',
    nav_product: 'ผลิตภัณฑ์',
    nav_solutions: 'โซลูชัน',
    nav_solutions_industry: 'ทุกอุตสาหกรรม',
    nav_solutions_enterprise: 'องค์กร (Enterprise)',
    nav_resources: 'แหล่งข้อมูล',
    nav_pricing: 'ราคา',
    nav_about: 'เกี่ยวกับเรา',
    nav_login: 'เข้าสู่ระบบ',
    nav_book_demo: 'นัดเดโม',
    nav_contact: 'ติดต่อ',
    nav_demo: 'เดโม',
    nav_open_menu: 'เปิดเมนู',
    lang_label: 'ภาษา',
    footer_blurb: 'ทีมงาน AI ศูนย์บริการสำหรับทีมซัพพอร์ตสมัยใหม่',
    footer_product: 'ผลิตภัณฑ์',
    footer_overview: 'ภาพรวม',
    footer_get_started: 'เริ่มต้นใช้งาน',
    footer_live_demo: 'เดโมสด',
    footer_register: 'สมัครใช้งาน',
    footer_contact_sales: 'ติดต่อฝ่ายขาย',
    footer_company: 'บริษัท',
    footer_demo_guide: 'คู่มือเดโม',
    footer_privacy: 'เก็บลีดอย่างเคารพความเป็นส่วนตัว · ไม่มีการชำระเงินบนหน้าการตลาด',
    footer_rights: 'Monti สงวนลิขสิทธิ์',
    footer_care: 'สร้างเพื่อทีมที่ตอบด้วยความใส่ใจ',

    home_title: 'Monti — บทสนทนา AI ที่เข้าใจคุณ',
    home_pill: '✦ เอเจนต์เสียง AI สำหรับธุรกิจยุคใหม่',
    home_h1_1: 'บทสนทนา AI',
    home_h1_2: 'ที่เข้าใจจริง',
    home_h1_3: 'ผลลัพธ์ที่สำคัญ',
    home_lede:
      'Monti คือทีมงานศูนย์บริการ AI ตลอด 24 ชั่วโมง เอเจนต์เสียงที่สื่อสารแบบมนุษย์ เข้าใจ ตอบ และแก้ปัญหา — ทุกช่องทางและทุกภาษา',
    home_cta_demo: 'ลองเดโมสด Monti',
    home_cta_video: 'ดูวิดีโอ',
    home_chip_voice: 'เอเจนต์เสียง AI',
    home_chip_lang: 'รองรับหลายภาษา',
    home_chip_km: 'ขับเคลื่อนด้วยความรู้',
    home_chip_secure: 'ปลอดภัยและสอดคล้องมาตรฐาน',
    home_ava_role: 'ซัพพอร์ตทั่วไป',
    home_ava_tone: 'อบอุ่นและอดทน',
    home_desk_title: 'Monti Caller Desk',
    home_desk_live: 'สด',
    home_desk_general: 'ทั่วไป',
    home_desk_billing: 'บิลลิ่ง',
    home_desk_tech: 'เทคนิค',
    home_desk_welcome: 'ยินดีต้อนรับสู่ศูนย์บริการ Monti วันนี้ให้ช่วยอะไรดีคะ?',
    home_desk_you: 'คุณ',
    home_desk_user_msg: 'ต้องการความช่วยเหลือเรื่องบิล',
    home_desk_just_now: 'เมื่อสักครู่',
    home_desk_placeholder: 'พิมพ์คำถามของคุณ…',
    home_desk_send: 'ส่ง',
    home_stat_60: 'ลดเวลาจัดการสาย',
    home_stat_1000s: 'รองรับสายพร้อมกันหลายพัน',
    home_stat_247: 'พร้อมเสมอ ไม่พลาดสาย',
    home_stat_csat: 'ลูกค้าพึงพอใจมากขึ้น',
    home_built_eyebrow: 'ออกแบบเพื่อธุรกิจ',
    home_built_h2: 'ทุกอย่างที่คุณต้องการเพื่อส่งมอบประสบการณ์ที่ดี',
    home_cap_voice_t: 'เสียง AI คล้ายมนุษย์',
    home_cap_voice_b: 'บทสนทนาธรรมชาติที่รู้สึกเหมือนคนจริง',
    home_cap_km_t: 'ความรู้ที่ฉลาด',
    home_cap_km_b: 'ตอบจากข้อมูล นโยบาย และเอกสารของคุณ',
    home_cap_omni_t: 'ทุกช่องทาง',
    home_cap_omni_b: 'เสียง เว็บวิดเจ็ต มือถือ และอื่นๆ',
    home_cap_workflow_t: 'พร้อมเชื่อมเวิร์กโฟลว์',
    home_cap_workflow_b: 'เชื่อมต่อกับระบบและเครื่องมือของคุณ',
    home_cap_handover_t: 'ส่งต่อคนจริง',
    home_cap_handover_b: 'โอนงานสู่เจ้าหน้าที่อย่างราบรื่น',
    home_cap_insights_t: 'ข้อมูลเชิงลึกและการวิเคราะห์',
    home_cap_insights_b: 'มอนิเตอร์แบบเรียลไทม์และอินไซต์ที่นำไปใช้ได้',
    home_use_eyebrow: 'กรณีการใช้งาน',
    home_use_h2: 'Monti ทำงานได้ทุกที่ที่คุณอยู่',
    home_use_1: 'ฝ่ายบริการลูกค้า',
    home_use_2: 'บิลลิ่งและการชำระเงิน',
    home_use_3: 'ซัพพอร์ตด้านเทคนิค',
    home_use_4: 'ขายและคัดกรองลีด',
    home_use_5: 'จองนัดหมาย',
    home_use_6: 'คีออสก์และบริการตนเอง',
    home_qr_h3: 'ลอง Monti จริง',
    home_qr_p: 'สัมผัสบทสนทนาเสียง AI สดในภาษาของคุณ',
    home_qr_cta: 'ลองเดโมสด',
    home_trust: 'ได้รับความไว้วางใจจากธุรกิจ',
    home_trust_note: 'ชื่อแบรนด์ตัวอย่างเพื่อเลย์เอาต์เท่านั้น — ไม่ใช่การรับรองจริง',
    home_ready_h2: 'พร้อมเมื่อคุณพร้อม',
    home_ready_p: 'ทดลองผลิตภัณฑ์สด พูดคุยกับฝ่ายขาย หรือเปิดเวิร์กสเปซและเลือกแพ็กเกจจากแคตตาล็อก',
    home_ready_demo: 'ลองเดโมสด',
    home_ready_book: 'นัดเดโม',
    home_ready_register: 'สมัครใช้งานฟรี',

    product_title: 'ผลิตภัณฑ์ — Monti Product Suite',
    product_h1: 'Monti Product Suite',
    product_lede: 'ทุกอย่างที่คุณต้องการเพื่อสร้าง รัน และขยายบทสนทนา AI ที่สร้างผลลัพธ์จริง',
    product_cta_demo: 'ลองเดโมสด',
    product_cta_register: 'สมัครใช้งานฟรี',
    product_suite_kicker: 'ชุดผลิตภัณฑ์',
    product_nav_overview: 'ภาพรวม',
    product_nav_voice: 'เอเจนต์เสียง AI',
    product_nav_omni: 'ทุกช่องทาง',
    product_nav_km: 'คลังความรู้',
    product_nav_handover: 'ส่งต่อคนจริง',
    product_nav_analytics: 'การวิเคราะห์และอินไซต์',
    product_nav_security: 'ความปลอดภัยและการปฏิบัติตาม',
    product_nav_integrations: 'การเชื่อมต่อระบบ',
    product_km_title: 'คลังความรู้',
    product_km_sub: 'คำตอบที่ยึดจากธุรกิจของคุณ',
    product_km_search: 'ค้นหาความรู้…',
    product_km_1: 'คำถามบิลลิ่งที่พบบ่อย',
    product_km_2: 'นโยบายคืนสินค้าและคืนเงิน',
    product_km_3: 'คู่มือทางเทคนิค',
    product_voice_talking: 'กำลังคุยกับ Ava',
    product_voice_role: 'เอเจนต์เสียง AI',
    product_voice_end: 'วางสาย',
    product_pitch_h2: 'ทรงพลังตามการออกแบบ',
    product_pitch_em: 'ใช้งานง่าย',
    product_pitch_p:
      'Monti รวม AI ขั้นสูงกับประสบการณ์ที่เข้าใจง่าย ให้ทีมโฟกัสสิ่งที่สำคัญที่สุด — ลูกค้าของคุณ',
    product_ben_easy_t: 'เริ่มต้นง่าย',
    product_ben_easy_b: 'พร้อมใช้งานในไม่กี่นาที ไม่ใช่หลายสัปดาห์',
    product_ben_scale_t: 'พร้อมขยาย',
    product_ben_scale_b: 'แพลตฟอร์มระดับองค์กร พร้อมเติบโตไปกับคุณ',
    product_ben_flex_t: 'ยืดหยุ่นและเปิด',
    product_ben_flex_b: 'ทำงานตามวิธีของคุณด้วยการเชื่อมต่อแบบเปิด',
    product_ben_secure_t: 'เชื่อถือได้และปลอดภัย',
    product_ben_secure_b: 'สร้างบนความปลอดภัย ความเป็นส่วนตัว และการปฏิบัติตาม',
    product_footer_eyebrow: 'เห็นภาพรวมทั้งหมด',
    product_footer_h2: 'ชุดเดียว ทุกบทสนทนา',
    product_footer_demo: 'สำรวจ Monti สด',
    product_footer_talk: 'คุยกับทีมเรา',

    solutions_title: 'โซลูชัน — Monti',
    solutions_h1: 'บทสนทนา AI สำหรับทุกอุตสาหกรรมและทุกทีม',
    solutions_p: 'Monti ปรับให้เข้ากับธุรกิจของคุณ และส่งมอบประสบการณ์ที่ดีขึ้นในทุกจุดสัมผัส',
    solutions_learn: 'เรียนรู้เพิ่ม',
    solutions_ind_cs_t: 'ฝ่ายบริการลูกค้า',
    solutions_ind_cs_b: 'แก้ปัญหาลูกค้าได้เร็วขึ้นด้วยเอเจนต์ AI ที่เข้าใจและตอบอย่างเป็นธรรมชาติ',
    solutions_ind_fin_t: 'บริการทางการเงิน',
    solutions_ind_fin_b: 'บทสนทนาที่ปลอดภัย เชื่อถือได้ และสอดคล้องมาตรฐานสำหรับธนาคาร ประกัน และฟินเทค',
    solutions_ind_ecom_t: 'อีคอมเมิร์ซ',
    solutions_ind_ecom_b: 'เพิ่มยอดขายและความพึงพอใจด้วยผู้ช่วยช้อปปิ้ง AI ตลอด 24 ชม.',
    solutions_ind_health_t: 'สุขภาพ',
    solutions_ind_health_b: 'ช่วยเหลือผู้ป่วยและนัดหมายด้วยความเข้าใจและแม่นยำ',
    solutions_ind_travel_t: 'ท่องเที่ยวและโรงแรม',
    solutions_ind_travel_b: 'ตอบคำถามแขก จัดการการจอง และส่งมอบประสบการณ์ที่ดี',
    solutions_ind_tel_t: 'โทรคมนาคม',
    solutions_ind_tel_b: 'ลด churn และยกระดับ CX ด้วยบริการตนเองและซัพพอร์ตอัจฉริยะ',
    solutions_ind_edu_t: 'การศึกษา',
    solutions_ind_edu_b: 'มีส่วนร่วมกับนักเรียนและผู้ปกครองด้วยบอทข้อมูลและซัพพอร์ต',
    solutions_ind_gov_t: 'ภาครัฐ',
    solutions_ind_gov_b: 'บริการประชาชนที่ดีขึ้นด้วย AI ที่ปลอดภัยและเข้าถึงได้',
    solutions_expert_h2: 'ยังไม่แน่ใจว่าจะเริ่มตรงไหน?',
    solutions_expert_p: 'ทีมของเราช่วยหาโซลูชันที่เหมาะกับธุรกิจของคุณได้',
    solutions_expert_cta: 'คุยกับผู้เชี่ยวชาญ',

    ent_title: 'โซลูชันองค์กร — Monti',
    ent_badge: 'โซลูชันสำหรับองค์กร',
    ent_h1_1: 'Enterprise AI พร้อม',
    ent_h1_2: 'Data Ownership',
    ent_h1_3: 'ในตัว',
    ent_lede:
      'ให้ Monti ดูแลประสบการณ์ AI ในขณะที่ข้อมูลการโทร ฐานความรู้ และข้อมูลธุรกิจของคุณอยู่บน on-premise หรือ private cloud ของคุณเอง',
    ent_cta_sales: 'คุยกับฝ่ายขาย',
    ent_cta_arch: 'ดูสถาปัตยกรรม',
    ent_layer_ai: 'ชั้นประสบการณ์ AI / AIAAS โดย MONTI',
    ent_layer_ai_sub: 'ดูแล อัปเดต และรักษาความปลอดภัยโดย Monti',
    ent_layer_data: 'ชั้นข้อมูลลูกค้า / สภาพแวดล้อมของคุณ',
    ent_layer_data_sub: 'ติดตั้ง จัดเก็บ และควบคุมโดยคุณ',
    ent_layer_env: 'ON-PREMISE / PRIVATE CLOUD / CUSTOMER VPC',
    ent_own_monti: 'ดูแลโดย Monti',
    ent_own_you: 'เป็นเจ้าของโดยคุณ',
    ent_secure_link: 'เชื่อมต่ออย่างปลอดภัยผ่านช่องทางเข้ารหัส',
    ent_trust_title: 'ข้อมูลของคุณยังเป็นของคุณ',
    ent_trust_1: 'เป็นเจ้าของบันทึกการโทร',
    ent_trust_2: 'เก็บฐานความรู้ในสภาพแวดล้อมของคุณ',
    ent_trust_3: 'ติดตั้ง on-premise หรือ private cloud',
    ent_trust_4: 'สถาปัตยกรรมพร้อม PDPA / องค์กร',
    ent_how_h2: 'วิธีการทำงาน',
    ent_how_1_t: '1. AI Frontend ที่จัดการให้',
    ent_how_1_b: 'Monti ส่งมอบประสบการณ์ AI — voice agents, chat UI, orchestration, guardrails และ analytics ในรูปแบบ managed service',
    ent_how_2_t: '2. Secure Connector / API Gateway',
    ent_how_2_b: 'ทุกการโต้ตอบผ่านเกตเวย์ที่ปลอดภัย พร้อมยืนยันตัวตน การเข้ารหัส และการควบคุมสิทธิ์',
    ent_how_3_t: '3. ชั้นข้อมูล On-Premise',
    ent_how_3_b: 'ชั้นข้อมูลอยู่ที่คุณ — บันทึกการโทร ฐานความรู้ ฐานข้อมูล และ audit logs',
    ent_how_4_t: '4. ตัวเลือกติดตั้งที่ยืดหยุ่น',
    ent_how_4_b: 'ติดตั้ง on-premise, private cloud หรือ hybrid ให้ตรงกับ IT และข้อกำหนด compliance ของคุณ',
    ent_managed_h2: 'ดูแลโดย MONTI',
    ent_managed_foot: 'Monti ดูแลประสบการณ์ AI เพื่อให้คุณโฟกัสกับธุรกิจ',
    ent_managed_1: 'AI Agents',
    ent_managed_2: 'Voice Interface',
    ent_managed_3: 'Orchestration',
    ent_managed_4: 'อัปเดตต่อเนื่อง',
    ent_managed_5: 'Support & Success',
    ent_controlled_h2: 'ควบคุมโดยคุณ',
    ent_controlled_foot: 'คุณเป็นเจ้าของข้อมูล นโยบาย และสิทธิ์เข้าถึง — ครบวงจร',
    ent_controlled_1: 'บันทึกการโทร',
    ent_controlled_2: 'เอกสาร KM',
    ent_controlled_3: 'ที่เก็บและฐานข้อมูล',
    ent_controlled_4: 'นโยบายเก็บรักษา',
    ent_controlled_5: 'การควบคุมสิทธิ์',
    ent_deploy_h2: 'ตัวเลือกการติดตั้ง',
    ent_deploy_1_t: 'On-Premise',
    ent_deploy_1_b: 'ติดตั้งชั้นข้อมูลใน data center ของคุณ พร้อมการควบคุมและแยกขาดเต็มรูปแบบ',
    ent_deploy_2_t: 'Private Cloud',
    ent_deploy_2_b: 'รันใน private cloud ของคุณด้วยทรัพยากรเฉพาะและความปลอดภัยระดับองค์กร',
    ent_deploy_3_t: 'Hybrid',
    ent_deploy_3_b: 'ผสมความปลอดภัย on-premise สำหรับข้อมูลสำคัญ กับความยืดหยุ่นของคลาวด์สำหรับงานอื่น',
    ent_ben_h2: 'ประโยชน์สำหรับองค์กร',
    ent_ben_1_t: 'Data Ownership',
    ent_ben_1_b: 'คุณเป็นเจ้าของและควบคุมข้อมูลทั้งหมด',
    ent_ben_2_t: 'ความปลอดภัยและ Compliance',
    ent_ben_2_b: 'ออกแบบตามมาตรฐานองค์กรและ PDPA',
    ent_ben_3_t: 'เร่งการเปิดใช้งาน',
    ent_ben_3_b: 'ประสบการณ์ AI สำเร็จรูป พร้อมติดตั้งเร็ว',
    ent_ben_4_t: 'สถาปัตยกรรมขยายได้',
    ent_ben_4_b: 'เติบโตไปพร้อมองค์กรของคุณ',
    ent_ben_5_t: 'พร้อมเชื่อมระบบ',
    ent_ben_5_b: 'เชื่อมกับระบบที่มีอยู่ของคุณ',
    ent_ben_6_t: 'ตรวจสอบย้อนหลังได้',
    ent_ben_6_b: 'มองเห็นครบด้วย audit logs และการติดตามสิทธิ์',
    ent_cta_h2: 'ออกแบบการติดตั้ง enterprise ที่เหมาะกับองค์กรของคุณ',
    ent_cta_p: 'ทีมเราจะทำงานร่วมกับ IT และ security ของคุณ เพื่อสร้างโมเดลการติดตั้งที่ตรงความต้องการและปกป้องสิ่งสำคัญที่สุด',
    ent_cta_demo: 'นัดเดโม enterprise',
    ent_cta_architect: 'คุยกับสถาปนิก',
    ent_arch_web: 'Web Widget',
    ent_arch_voice: 'Voice Agent',
    ent_arch_chat: 'Chat UI',
    ent_arch_orch: 'Orchestration',
    ent_arch_guard: 'Guardrails',
    ent_arch_analytics: 'Analytics Dashboard',
    ent_arch_calls: 'บันทึกการโทร',
    ent_arch_km: 'ฐานความรู้ (KM)',
    ent_arch_customer: 'ข้อมูลลูกค้า',
    ent_arch_storage: 'ที่เก็บข้อมูล',
    ent_arch_db: 'ฐานข้อมูล',
    ent_arch_audit: 'Audit Logs',

    resources_title: 'Embed และ Mobile SDK — Monti',
    resources_h1: 'Embed และ Mobile SDK',
    resources_p:
      'ฝัง Monti บนเว็บไซต์ด้วย SDK ตามเฟรมเวิร์ก และบนมือถือด้วยแบรนด์ของคุณ ให้ลูกค้าค้นหาผู้เช่าและโทรหา AI ได้โดยตรง',
    resources_web_eyebrow: 'เว็บฝัง',
    resources_web_h2: 'SDK เว็บไซต์แยกตามเทคโนโลยี',
    resources_install: 'ติดตั้ง',
    resources_example: 'ตัวอย่างโค้ดฝัง',
    resources_copy: 'คัดลอกโค้ด',
    resources_copied: 'คัดลอกแล้ว',
    resources_hint:
      'ต้องมี origin ของเซิร์ฟเวอร์ Monti (api-base / apiBase) และคีย์ฝังที่เปิดใช้ พร้อม allowlist ของโดเมนโฮสต์',
    resources_mobile_eyebrow: 'มือถือ · แบรนด์ของคุณ',
    resources_mobile_h2: 'แอปมือถือแบรนด์คุณเอง สำหรับลูกค้าโทรตรง',
    resources_mobile_p:
      'นอกจากฝังบนเว็บ Monti รองรับมือถือที่ลูกค้าเห็นแบรนด์คุณ: เรียกดูรายชื่อผู้เช่า สร้างแบรนด์ เลือกเอเจนต์ AI แล้วเริ่มแชทหรือโทรเสียงได้ทันที — ชื่อบริษัท เอวาตาร์ และภาษา (EN/TH/JA)',
    resources_mobile_f1_t: 'ไดเรกทอรีแบรนด์',
    resources_mobile_f1_b: 'ลูกค้าค้นหาและเปิดแบรนด์ผู้เช่าที่เปิดเผยสาธารณะจากแอปมือถือ',
    resources_mobile_f2_t: 'พื้นผิวแบรนด์ของคุณ',
    resources_mobile_f2_b: 'ชื่อผู้เช่า เอเจนต์ ภาษา และโควตา แสดงภายใต้แบรนด์คุณก่อนเริ่มสาย',
    resources_mobile_f3_t: 'โทร AI โดยตรง',
    resources_mobile_f3_b: 'เริ่มโทรหรือแชทจากมือถือ — เอวาตาร์แบรนด์ เสียงสด และทรานสคริปต์',
    resources_mobile_cap_brands: 'ค้นหาแบรนด์ / ผู้เช่า',
    resources_mobile_cap_tenant: 'แบรนด์ผู้เช่า + เอเจนต์ AI',
    resources_mobile_cap_call: 'สายสดพร้อมแบรนด์',
    resources_mobile_sdk_h3: 'ตัวอย่าง Mobile SDK',
    resources_mobile_sdk_p:
      'คอร์ TypeScript สำหรับโฮสต์ iOS, Android, React Native และ Flutter การยืนยันตัวตน เอวาตาร์ โควตา และวงจรสายอยู่ฝั่งเซิร์ฟเวอร์',

    pricing_title: 'ราคา — Monti',
    pricing_h1: 'ราคาเรียบง่าย โปร่งใส',
    pricing_p: 'เลือกแผนที่เหมาะกับธุรกิจของคุณ อัปเกรดหรือดาวน์เกรดได้ทุกเมื่อ',
    pricing_monthly: 'รายเดือน',
    pricing_annual: 'รายปี',
    pricing_save: 'ประหยัด 20%',
    pricing_loading: 'กำลังโหลดแพ็กเกจ…',
    pricing_error: 'แคตตาล็อกแพ็กเกจใช้ชั่วคราวไม่ได้',
    pricing_error_hint: 'แสดงแผนอ้างอิง — ยืนยันราคาจริงหลังสมัครหรือติดต่อฝ่ายขาย',
    pricing_most_popular: 'ยอดนิยม',
    pricing_choose: 'เลือก',
    pricing_contact_sales: 'ติดต่อฝ่ายขาย',
    pricing_catalog_note: 'ราคาจริงจากแคตตาล็อก · สิทธิ์ใช้งานเริ่มหลังชำระเงินเท่านั้น',
    pricing_no_setup_t: 'ไม่มีค่าติดตั้ง',
    pricing_no_setup_b: 'เริ่มได้ในไม่กี่นาที',
    pricing_cancel_t: 'ยกเลิกได้ทุกเมื่อ',
    pricing_cancel_b: 'ไม่มีสัญญาผูกมัดยาว',
    pricing_secure_t: 'ปลอดภัยและสอดคล้องมาตรฐาน',
    pricing_secure_b: 'ความปลอดภัยระดับองค์กร',
    pricing_support_t: 'ซัพพอร์ตเฉพาะทาง',
    pricing_support_b: 'เราพร้อมช่วยเหลือคุณ',
    pricing_legal:
      'นาทีการโทรและการใช้งานอยู่ภายใต้นโยบายการใช้งานอย่างเป็นธรรม ราคายังไม่รวมภาษี ราคาบนหน้าสาธารณะไม่ให้โควตา — การชำระเงินอยู่ในบิลลิ่งของผู้เช่าที่ล็อกอินแล้ว',
    pricing_perfect_start: 'เหมาะสำหรับเริ่มต้น',
    pricing_growing: 'สำหรับธุรกิจที่กำลังเติบโต',
    pricing_scaling: 'สำหรับทีมที่ขยายตัว',
    pricing_large: 'สำหรับองค์กรขนาดใหญ่',
    pricing_custom: 'กำหนดเอง',
    pricing_blurb_starter: 'เหมาะสำหรับเริ่มต้น',
    pricing_blurb_growth: 'สำหรับธุรกิจที่กำลังเติบโต',
    pricing_blurb_pro: 'สำหรับทีมที่ขยายตัว',
    pricing_blurb_ent: 'สำหรับองค์กรขนาดใหญ่',
    pricing_shared_h2: 'Shared Cloud — สตาร์ทอัพและ SME',
    pricing_shared_p:
      'แพลตฟอร์มมัลติ-เทแนนท์ ซื้อออนไลน์ผ่านช่องทางชำระเงิน ใช้ API key AI ของคุณเอง นาทีเสียงบนแพลตฟอร์มไม่จำกัด',
    pricing_dedicated_h2: 'Dedicated VM — ขอใบเสนอราคา',
    pricing_dedicated_p:
      'โครงสร้างแยกสำหรับองค์กรที่ต้องการความเป็นส่วนตัว ขอใบเสนอราคาเพื่อตรวจความพร้อมของเซิร์ฟเวอร์ก่อนจัดสรร',
    pricing_request_quote: 'ขอใบเสนอราคา',
    pricing_buy_now: 'ซื้อพร้อมชำระเงิน',

    about_title: 'เกี่ยวกับเรา — Monti',
    about_h1: 'เราสร้าง AI ที่รู้สึกเหมือนบทสนทนาของมนุษย์',
    about_p1: 'บทสนทนาที่เป็นธรรมชาติ แก้ปัญหาได้ฉลาดขึ้น และสร้างความสัมพันธ์ที่แข็งแรงขึ้น',
    about_p2:
      'ทีมของเรารวมความเชี่ยวชาญด้าน AI เทคโนโลยีเสียง และประสบการณ์ลูกค้า เพื่อสร้างโซลูชันที่สร้างความแตกต่างจริง',
    about_story: 'ดูบทสนทนา — เปิดเดโมสด',
    about_stat_team: 'สมาชิกทีม',
    about_stat_customers: 'ลูกค้าที่พึงพอใจ',
    about_stat_convos: 'บทสนทนาต่อเดือน',
    about_stat_support: 'ซัพพอร์ตและความสำเร็จ',
    about_values: 'ทำไมต้อง Monti',
    about_val_customer_t: 'Natural Voice AI',
    about_val_customer_b: 'บทสนทนาเหมือนมนุษย์ที่สร้างความไว้ใจ',
    about_val_innov_t: 'ฉลาดและมีประสิทธิภาพ',
    about_val_innov_b: 'แก้ปัญหาได้มากขึ้นด้วยบริบท ไม่ใช่สคริปต์',
    about_val_integ_t: 'Human + AI',
    about_val_integ_b: 'ส่งต่องานไร้รอยต่อ แข็งแกร่งกว่าเมื่อทำงานร่วมกัน',
    about_val_impact_t: 'ปลอดภัยและสอดคล้องมาตรฐาน',
    about_val_impact_b: 'ความปลอดภัยระดับองค์กรที่คุณวางใจได้',
    about_cta_h2: 'มาร่วมสร้างบทสนทนาที่ดีกว่าด้วยกัน',
    about_cta_p: 'เรายินดีรับฟังความท้าทายของคุณ และสำรวจว่า Monti ช่วยธุรกิจคุณเติบโตได้อย่างไร',
    about_cta_btn: 'นัดเดโม',

    contact_title: 'ติดต่อ — Monti',
    contact_book: 'นัดเดโม',
    contact_sales: 'ติดต่อฝ่ายขาย',
    contact_newsletter: 'ติดตามข่าวสาร',
    contact_lede:
      'บอกเราเกี่ยวกับเป้าหมายซัพพอร์ตขาเข้าของคุณ เราติดต่อกลับเฉพาะเมื่อคุณยินยอม ไม่มีการชำระเงินในหน้านี้',
    contact_thanks: 'ขอบคุณ — ฝ่ายขายจะติดต่อกลับ',
    contact_deduped: 'เรามีคำขอล่าสุดจากอีเมลนี้สำหรับความตั้งใจเดียวกันแล้ว',
    contact_received: 'ได้รับคำขอแล้ว ทีมจะติดต่อผ่านช่องทางที่คุณเลือกเมื่อพร้อม',
    contact_see_demo: 'ดูตัวเลือกเดโมสด',
    contact_back_home: 'กลับหน้าแรก',
    contact_type_book: 'นัดเดโม',
    contact_type_contact: 'ติดต่อ',
    contact_name: 'ชื่อ-นามสกุล',
    contact_email: 'อีเมลที่ทำงาน',
    contact_company: 'บริษัท',
    contact_phone: 'โทรศัพท์',
    contact_usecase: 'กรณีการใช้งาน',
    contact_channel: 'ช่องทางที่สะดวก',
    contact_channel_email: 'อีเมล',
    contact_channel_phone: 'โทรศัพท์',
    contact_channel_line: 'LINE',
    contact_channel_other: 'อื่นๆ',
    contact_consent_contact: 'ฉันยินยอมให้ติดต่อเกี่ยวกับผลิตภัณฑ์และเดโมของ Monti',
    contact_consent_marketing: 'ฉันต้องการรับอัปเดตผลิตภัณฑ์และอีเมลการตลาด (ไม่บังคับ)',
    contact_submit: 'ส่ง',
    contact_submitting: 'กำลังส่ง…',
    contact_err_email: 'ต้องระบุอีเมล',
    contact_err_consent: 'ต้องยินยอมการติดต่อ',
    contact_err_marketing: 'ต้องยินยอมการตลาดสำหรับสมัครรับข่าว',
    contact_err_generic: 'ส่งไม่สำเร็จ กรุณาลองใหม่ในอีกสักครู่',

    demo_title: 'เดโมสด — Monti',
    demo_badge: 'เดโมสด',
    demo_h1: 'สัมผัสเอวาตาร์ AI ของ Monti โดยไม่ต้องมีบัญชี',
    demo_lede:
      'เดโมสดคือพอร์ทัลลูกค้าแบบไม่ล็อกอิน เลือกเอเจนต์เอวาตาร์ AI แล้วถามด้วยข้อความหรือเสียง พารามิเตอร์การอ้างอิงจากไซต์ผลิตภัณฑ์จะถูกเก็บต่อเมื่อคุณไปต่อ',
    demo_try_h2: 'สิ่งที่ลองได้',
    demo_try_1: 'เลือกเอเจนต์เอวาตาร์ AI ตามแบรนด์',
    demo_try_2: 'ถามคำถามแบบผลิตภัณฑ์ด้วยข้อความ',
    demo_try_3: 'โต้ตอบด้วยเสียงเมื่อเปิดใช้งาน',
    demo_try_4: 'สัมผัสว่า AI ขาเข้าเป็นอย่างไรสำหรับผู้เยี่ยมชม',
    demo_after_h2: 'หลังเดโม',
    demo_after_1: 'นัดวอล์กทรูกับฝ่ายขาย',
    demo_after_2: 'สมัครเวิร์กสเปซผู้เช่า',
    demo_after_3: 'เลือกแพ็กเกจจากแคตตาล็อกจริง',
    demo_after_4: 'การชำระเงินยังอยู่ในบิลลิ่งที่ล็อกอินแล้ว',
    demo_open_h2: 'เปิดเดโมสด',
    demo_open_p: 'คุณจะออกจากไซต์การตลาดไปยังพื้นผิวเดโมลูกค้าที่รากของไซต์',
    demo_launch: 'เปิดเดโมสด',
    demo_book: 'นัดเดโมพร้อมไกด์',
    demo_register: 'เริ่มสมัครใช้งาน'
  },

  ja: {
    brand_tagline: 'AIコールセンター',
    nav_product: 'プロダクト',
    nav_solutions: 'ソリューション',
    nav_solutions_industry: 'あらゆる業界',
    nav_solutions_enterprise: 'エンタープライズ',
    nav_resources: 'リソース',
    nav_pricing: '料金',
    nav_about: '会社情報',
    nav_login: 'ログイン',
    nav_book_demo: 'デモを予約',
    nav_contact: 'お問い合わせ',
    nav_demo: 'デモ',
    nav_open_menu: 'メニューを開く',
    lang_label: '言語',
    footer_blurb: '現代のサポートチームのためのAIコールセンター人材。',
    footer_product: 'プロダクト',
    footer_overview: '概要',
    footer_get_started: 'はじめる',
    footer_live_demo: 'ライブデモ',
    footer_register: '登録',
    footer_contact_sales: '営業に連絡',
    footer_company: '会社',
    footer_demo_guide: 'デモガイド',
    footer_privacy: 'プライバシーに配慮したリード取得 · マーケティングページでは決済なし',
    footer_rights: 'Monti. All rights reserved.',
    footer_care: '丁寧な対応を大切にするチームのために。',
    home_title: 'Monti — 理解するAI会話',
    home_pill: '✦ 現代ビジネスのためのAI音声エージェント',
    home_h1_1: 'AI会話が',
    home_h1_2: '理解する。',
    home_h1_3: '成果が残る。',
    home_lede: 'Montiは24時間稼働のAIコールセンター人材です。人間らしい音声エージェントが理解し、応答し、解決します — あらゆるチャネルと言語で。',
    home_cta_demo: 'Montiライブデモを試す',
    home_cta_video: '動画を見る',
    home_chip_voice: '音声AIエージェント',
    home_chip_lang: '多言語対応',
    home_chip_km: 'ナレッジ連携',
    home_chip_secure: '安全・コンプライアンス',
    home_ava_role: '一般サポート',
    home_ava_tone: '温かく丁寧',
    home_desk_title: 'Montiコーラーデスク',
    home_desk_live: 'Live',
    home_desk_general: '一般',
    home_desk_billing: '請求',
    home_desk_tech: '技術',
    home_desk_welcome: 'Montiインバウンドコールセンターへようこそ。本日はどのようなご用件でしょうか？',
    home_desk_you: 'あなた',
    home_desk_user_msg: '請求についてサポートが必要です。',
    home_desk_just_now: 'たった今',
    home_desk_placeholder: '質問を入力…',
    home_desk_send: '送信',
    home_stat_60: '通話処理時間を短縮',
    home_stat_1000s: '同時に対応できる通話数',
    home_stat_247: '常時稼働、見逃しゼロ',
    home_stat_csat: 'より良い顧客体験',
    home_built_eyebrow: 'ビジネスのために設計',
    home_built_h2: '優れた体験に必要なすべて',
    home_cap_voice_t: '人間らしいAI音声',
    home_cap_voice_b: '人に近い自然な会話',
    home_cap_km_t: 'スマートナレッジ',
    home_cap_km_b: 'データ・方針・ドキュメントに基づく回答',
    home_cap_omni_t: 'オムニチャネル',
    home_cap_omni_b: '音声、Webウィジェット、モバイルなど',
    home_cap_workflow_t: 'ワークフロー対応',
    home_cap_workflow_b: '既存のシステムとツールに連携',
    home_cap_handover_t: 'ライブ引き継ぎ',
    home_cap_handover_b: '有人オペレーターへスムーズに転送',
    home_cap_insights_t: 'インサイトと分析',
    home_cap_insights_b: 'リアルタイム監視と実行可能な洞察',
    home_use_eyebrow: 'ユースケース',
    home_use_h2: 'Montiはあなたの現場で働きます',
    home_use_1: 'カスタマーサポート',
    home_use_2: '請求と支払い',
    home_use_3: 'テクニカルサポート',
    home_use_4: '営業・リード資格判定',
    home_use_5: '予約受付',
    home_use_6: 'キオスク・セルフサービス',
    home_qr_h3: 'Montiの動きを見る',
    home_qr_p: 'あなたの言語でライブAI音声会話を体験。',
    home_qr_cta: 'ライブデモを試す',
    home_trust: '多くの企業に信頼されています',
    home_trust_note: 'レイアウト用のプレースホルダーブランド — 実際の推奨ではありません。',
    home_ready_h2: 'いつでも準備完了',
    home_ready_p: '製品をライブで体験、営業に相談、またはテナントワークスペースを開いてカタログからプランを選択。',
    home_ready_demo: 'ライブデモを試す',
    home_ready_book: 'デモを予約',
    home_ready_register: '無料登録を開始',
    product_title: 'プロダクト — Montiプロダクトスイート',
    product_h1: 'Montiプロダクトスイート',
    product_lede: '成果を生むAI会話を構築・運用・拡張するために必要なすべて。',
    product_cta_demo: 'ライブデモを試す',
    product_cta_register: '無料登録を開始',
    product_suite_kicker: 'プロダクトスイート',
    product_nav_overview: '概要',
    product_nav_voice: 'AI音声エージェント',
    product_nav_omni: 'オムニチャネル',
    product_nav_km: 'ナレッジハブ',
    product_nav_handover: 'ライブ引き継ぎ',
    product_nav_analytics: '分析とインサイト',
    product_nav_security: 'セキュリティとコンプライアンス',
    product_nav_integrations: '連携',
    product_km_title: 'ナレッジハブ',
    product_km_sub: 'ビジネスに根ざした回答',
    product_km_search: 'ナレッジを検索…',
    product_km_1: '請求FAQ',
    product_km_2: '返品・返金ポリシー',
    product_km_3: '技術ガイド',
    product_voice_talking: 'Avaと通話中',
    product_voice_role: 'AI音声エージェント',
    product_voice_end: '通話終了',
    product_pitch_h2: '設計はパワフル。',
    product_pitch_em: '使い方はシンプル。',
    product_pitch_p: 'Montiは高度なAIと直感的な体験を組み合わせ、チームが最も大切なこと — お客様 — に集中できるようにします。',
    product_ben_easy_t: 'すぐに始められる',
    product_ben_easy_b: '数週間ではなく数分で稼働。',
    product_ben_scale_t: 'スケール対応',
    product_ben_scale_b: '成長に合わせて拡張できるエンタープライズ級プラットフォーム。',
    product_ben_flex_t: '柔軟でオープン',
    product_ben_flex_b: 'オープンな連携で、あなたのやり方に合わせる。',
    product_ben_secure_t: '信頼性とセキュリティ',
    product_ben_secure_b: 'セキュリティ、プライバシー、コンプライアンスを中核に構築。',
    product_footer_eyebrow: '全体像を見る',
    product_footer_h2: 'ひとつのスイート。すべての会話。',
    product_footer_demo: 'Montiをライブで探索',
    product_footer_talk: 'チームに相談',
    solutions_title: 'ソリューション — Monti',
    solutions_h1: 'あらゆる業界とチームのためのAI会話',
    solutions_p: 'Montiはビジネスニーズに適応し、すべての接点でより良い体験を届けます。',
    solutions_learn: '詳しく見る',
    solutions_ind_cs_t: 'カスタマーサポート',
    solutions_ind_cs_b: '自然に理解し応答するAIエージェントで、顧客課題をより早く解決。',
    solutions_ind_fin_t: '金融サービス',
    solutions_ind_fin_b: '銀行・保険・フィンテック向けの安全で信頼性の高い会話。',
    solutions_ind_ecom_t: 'Eコマース',
    solutions_ind_ecom_b: '24時間のAIショッピングアシスタントで売上と満足度を向上。',
    solutions_ind_health_t: 'ヘルスケア',
    solutions_ind_health_b: '共感と正確さで患者サポートと予約案内を提供。',
    solutions_ind_travel_t: '旅行・ホスピタリティ',
    solutions_ind_travel_b: 'ゲストの質問に答え、予約を管理し、心地よい体験を提供。',
    solutions_ind_tel_t: '通信',
    solutions_ind_tel_b: '賢いセルフサービスとサポートでチャーンを削減しCXを改善。',
    solutions_ind_edu_t: '教育',
    solutions_ind_edu_b: 'スマートな情報・サポートボットで学生と保護者に関与。',
    solutions_ind_gov_t: '公共セクター',
    solutions_ind_gov_b: '安全でアクセスしやすいAI対話で市民サービスを向上。',
    solutions_expert_h2: 'どこから始めればよいか分からない？',
    solutions_expert_p: 'ビジネスに合うソリューションをチームがご案内します。',
    solutions_expert_cta: '専門家に相談',
    ent_title: 'エンタープライズソリューション — Monti',
    ent_badge: 'エンタープライズ向けソリューション',
    ent_h1_1: 'エンタープライズAIに',
    ent_h1_2: 'データ所有権を',
    ent_h1_3: '組み込み',
    ent_lede: 'AI体験はMontiが管理しつつ、通話記録・ナレッジベース・業務データはオンプレミスまたはプライベートクラウド環境に配置できます。',
    ent_cta_sales: '営業に相談',
    ent_cta_arch: 'アーキテクチャを見る',
    ent_layer_ai: 'AI体験レイヤー / AIAAS BY MONTI',
    ent_layer_ai_sub: 'Montiが管理・更新・保護',
    ent_layer_data: '顧客データレイヤー / お客様の環境',
    ent_layer_data_sub: 'お客様が配置・保存・制御',
    ent_layer_env: 'オンプレミス / プライベートクラウド / 顧客VPC',
    ent_own_monti: 'Montiが管理',
    ent_own_you: 'お客様が所有',
    ent_secure_link: '暗号化チャネルによる安全な接続',
    ent_trust_title: 'データはお客様のものです。',
    ent_trust_1: '通話記録を所有',
    ent_trust_2: 'ナレッジベースを自社環境に保持',
    ent_trust_3: 'オンプレミスまたはプライベートクラウドに配置',
    ent_trust_4: 'PDPA / エンタープライズ対応アーキテクチャ',
    ent_how_h2: '仕組み',
    ent_how_1_t: '1. マネージドAIフロントエンド',
    ent_how_1_b: 'Montiが音声エージェント、チャットUI、オーケストレーション、ガードレール、分析をマネージドサービスとして提供。',
    ent_how_2_t: '2. セキュアコネクタ / APIゲートウェイ',
    ent_how_2_b: 'すべての対話は強力な認証・暗号化・アクセス制御を備えた安全なゲートウェイを通過。',
    ent_how_3_t: '3. オンプレミスデータレイヤー',
    ent_how_3_b: 'データレイヤーはお客様環境に — 通話記録、ナレッジベース、データベース、監査ログ。',
    ent_how_4_t: '4. 柔軟な導入とセットアップ',
    ent_how_4_b: 'オンプレミス、プライベートクラウド、またはハイブリッド — ITとコンプライアンス要件に合わせて。',
    ent_managed_h2: 'MONTIが管理',
    ent_managed_foot: 'MontiがAI体験を管理するので、ビジネスに集中できます。',
    ent_managed_1: 'AIエージェント',
    ent_managed_2: '音声インターフェース',
    ent_managed_3: 'オーケストレーション',
    ent_managed_4: '継続的な更新',
    ent_managed_5: 'サポートと成功',
    ent_controlled_h2: 'お客様が制御',
    ent_controlled_foot: 'データ、ポリシー、アクセスをエンドツーエンドで所有。',
    ent_controlled_1: '通話記録',
    ent_controlled_2: 'KMドキュメント',
    ent_controlled_3: 'ストレージとデータベース',
    ent_controlled_4: '保持ポリシー',
    ent_controlled_5: 'アクセス制御',
    ent_deploy_h2: '導入オプション',
    ent_deploy_1_t: 'オンプレミス',
    ent_deploy_1_b: '自社データセンター内にデータレイヤーを配置し、完全な制御と隔離を実現。',
    ent_deploy_2_t: 'プライベートクラウド',
    ent_deploy_2_b: '専用リソースとエンタープライズ級セキュリティでプライベートクラウド環境に実行。',
    ent_deploy_3_t: 'ハイブリッド',
    ent_deploy_3_b: '機密データはオンプレミスのセキュリティ、その他はクラウドの柔軟性を組み合わせ。',
    ent_ben_h2: 'エンタープライズの利点',
    ent_ben_1_t: 'データ所有権',
    ent_ben_1_b: 'すべてのデータを所有・制御',
    ent_ben_2_t: 'セキュリティとコンプライアンス',
    ent_ben_2_b: 'エンタープライズ基準とPDPA向けに構築',
    ent_ben_3_t: '迅速なロールアウト',
    ent_ben_3_b: '事前構築のAI体験で素早く導入',
    ent_ben_4_t: 'スケーラブルなアーキテクチャ',
    ent_ben_4_b: '組織の成長に合わせて拡張',
    ent_ben_5_t: '連携対応',
    ent_ben_5_b: '既存システムと接続',
    ent_ben_6_t: '監査可能性',
    ent_ben_6_b: '監査ログとアクセス追跡で完全な可視性',
    ent_cta_h2: '組織に最適なエンタープライズ導入を一緒に設計しましょう。',
    ent_cta_p: 'ITとセキュリティチームと連携し、要件を満たし大切なものを守る導入モデルを作成します。',
    ent_cta_demo: 'エンタープライズデモを予約',
    ent_cta_architect: 'アーキテクトに相談',
    ent_arch_web: 'Webウィジェット',
    ent_arch_voice: '音声エージェント',
    ent_arch_chat: 'チャットUI',
    ent_arch_orch: 'オーケストレーション',
    ent_arch_guard: 'ガードレール',
    ent_arch_analytics: '分析ダッシュボード',
    ent_arch_calls: '通話記録',
    ent_arch_km: 'ナレッジベース (KM)',
    ent_arch_customer: '顧客データ',
    ent_arch_storage: 'ストレージ',
    ent_arch_db: 'データベース',
    ent_arch_audit: '監査ログ',
    resources_title: '埋め込み & モバイルSDK — Monti',
    resources_h1: '埋め込み & モバイルSDK',
    resources_p: 'フレームワーク埋め込みでWebサイトに、モバイルでは自社ブランドで顧客がテナントを検索しAIに直接通話できる体験を提供。',
    resources_web_eyebrow: 'Web埋め込み',
    resources_web_h2: '技術別WebサイトSDK',
    resources_install: 'インストール',
    resources_example: '埋め込みコード例',
    resources_copy: 'コードをコピー',
    resources_copied: 'コピーしました',
    resources_hint: 'Montiサーバーorigin（api-base / apiBase）と、ホストoriginを許可した有効な埋め込みキーが必要です。',
    resources_mobile_eyebrow: 'モバイル · 自社ブランド',
    resources_mobile_h2: '顧客直接通話向けカスタムブランドモバイルアプリ',
    resources_mobile_p: 'Web埋め込みに加え、顧客がブランド付き体験を開き、公開テナントブランドを閲覧し、AIエージェントを選び、チャットまたは音声通話を直接開始できるモバイルパスを提供 — 会社名、アバター、ロケール（EN/TH/JA）。',
    resources_mobile_f1_t: 'ブランドディレクトリ',
    resources_mobile_f1_b: '顧客は公開モバイルディレクトリから掲載テナントブランドを検索・開始。',
    resources_mobile_f2_t: '自社ブランド画面',
    resources_mobile_f2_b: '通話開始前にテナント名、エージェント、言語、クォータが自社ブランドで表示。',
    resources_mobile_f3_t: '直接AI通話',
    resources_mobile_f3_b: 'モバイルから通話またはチャットを開始 — ブランドアバター、ライブ音声、文字起こし。',
    resources_mobile_cap_brands: 'ブランド / テナントを探す',
    resources_mobile_cap_tenant: 'テナントブランド + AIエージェント',
    resources_mobile_cap_call: 'ライブブランド通話',
    resources_mobile_sdk_h3: 'モバイルSDK例',
    resources_mobile_sdk_p: 'iOS、Android、React Native、Flutterホスト向けTypeScriptコア。認証、アバター、クォータ、通話ライフサイクルはサーバー側で管理。',
    pricing_title: '料金 — Monti',
    pricing_h1: 'シンプルで透明な料金',
    pricing_p: 'ビジネスに合うプランを選択。いつでもアップグレード・ダウングレード可能。',
    pricing_monthly: '月額',
    pricing_annual: '年額',
    pricing_save: '20%お得',
    pricing_loading: 'パッケージを読み込み中…',
    pricing_error: 'パッケージカタログは一時的に利用できません。',
    pricing_error_hint: '参考プランを表示中 — 登録後にライブ価格を確認するか、営業にお問い合わせください。',
    pricing_most_popular: '最も人気',
    pricing_choose: '選択',
    pricing_contact_sales: '営業に連絡',
    pricing_catalog_note: 'パッケージカタログのライブ価格 · チェックアウト後にのみエンタイトルメント開始',
    pricing_no_setup_t: '初期費用なし',
    pricing_no_setup_b: '数分で開始。',
    pricing_cancel_t: 'いつでも解約',
    pricing_cancel_b: '長期契約なし。',
    pricing_secure_t: '安全・準拠',
    pricing_secure_b: 'エンタープライズ級セキュリティ。',
    pricing_support_t: '専任サポート',
    pricing_support_b: 'いつでもお手伝いします。',
    pricing_legal: '通話分数と利用は公正利用ポリシーの対象です。価格は税抜。公開料金だけではクォータは付与されません — 支払いは認証済みテナント請求で行います。',
    pricing_perfect_start: 'はじめに最適',
    pricing_growing: '成長企業向け',
    pricing_scaling: '拡大中のチーム向け',
    pricing_large: '大規模組織向け',
    pricing_custom: 'カスタム',
    pricing_blurb_starter: 'はじめに最適',
    pricing_blurb_growth: '成長企業向け',
    pricing_blurb_pro: '拡大中のチーム向け',
    pricing_blurb_ent: '大規模組織向け',
    pricing_shared_h2: '共有クラウド — スタートアップ & SME',
    pricing_shared_p: 'マルチテナントプラットフォーム。決済ゲートウェイでオンライン購入。AI APIキーはBYOK。プラットフォーム音声分数は無制限。',
    pricing_dedicated_h2: '専用VM — 見積もり依頼',
    pricing_dedicated_p: '大規模組織向け隔離インフラ。プロビジョニング前にサーバー容量を確認するため見積もりをご依頼ください。',
    pricing_request_quote: '見積もりを依頼',
    pricing_buy_now: '決済して購入',
    about_title: '会社情報 — Monti',
    about_h1: '人間の会話のように感じるAIを構築',
    about_p1: '自然な会話。より賢い解決。より強い関係。',
    about_p2: 'AI、音声技術、顧客体験の深い専門性を結集し、本当に差がつくソリューションを構築しています。',
    about_story: '会話を見る — ライブデモを開く',
    about_stat_team: 'チームメンバー',
    about_stat_customers: '満足したお客様',
    about_stat_convos: '月間対応会話数',
    about_stat_support: 'サポートと成功',
    about_values: 'Montiが選ばれる理由',
    about_val_customer_t: 'Natural Voice AI',
    about_val_customer_b: '信頼を築く人間らしい会話。',
    about_val_innov_t: '効率的でスマート',
    about_val_innov_b: 'スクリプトではなく文脈でより多く解決。',
    about_val_integ_t: 'Human + AI',
    about_val_integ_b: 'シームレスな引き継ぎ。一緒により強く。',
    about_val_impact_t: '安全・準拠',
    about_val_impact_b: '信頼できるエンタープライズ級セキュリティ。',
    about_cta_h2: 'より良い会話を一緒に作りましょう。',
    about_cta_p: '課題をお聞かせください。Montiがビジネス成長にどう役立つかご一緒に検討します。',
    about_cta_btn: 'デモを予約',
    contact_title: 'お問い合わせ — Monti',
    contact_book: 'デモを予約',
    contact_sales: '営業に連絡',
    contact_newsletter: '最新情報を受け取る',
    contact_lede: 'インバウンドサポートの目標をお知らせください。同意いただいた場合のみフォローアップします。このページでは決済は行いません。',
    contact_thanks: 'ありがとうございます — 営業よりご連絡します。',
    contact_deduped: '同じ意図でこのメールからの最近のリクエストが既にあります。',
    contact_received: 'リクエストを受け付けました。ご希望のチャネルで担当者から連絡します。',
    contact_see_demo: 'ライブデモの選択肢を見る',
    contact_back_home: 'ホームに戻る',
    contact_type_book: 'デモを予約',
    contact_type_contact: 'お問い合わせ',
    contact_name: '氏名',
    contact_email: '勤務先メール',
    contact_company: '会社名',
    contact_phone: '電話番号',
    contact_usecase: 'ユースケース',
    contact_channel: '希望チャネル',
    contact_channel_email: 'メール',
    contact_channel_phone: '電話',
    contact_channel_line: 'LINE',
    contact_channel_other: 'その他',
    contact_consent_contact: 'Monti製品とデモについて連絡を受けることに同意します。',
    contact_consent_marketing: '製品アップデートとマーケティングメールを希望（任意）。',
    contact_submit: '送信',
    contact_submitting: '送信中…',
    contact_err_email: 'メールは必須です。',
    contact_err_consent: '連絡への同意が必要です。',
    contact_err_marketing: 'ニュースレター登録にはマーケティング同意が必要です。',
    contact_err_generic: '送信できませんでした。しばらくしてから再度お試しください。',
    demo_title: 'ライブデモ — Monti',
    demo_badge: 'ライブデモ',
    demo_h1: 'アカウントなしでMonti AIアバターを体験',
    demo_lede: 'ライブデモは既存の認証なし顧客ポータルです。AIアバターエージェントを選び、テキストまたは音声で質問できます。本製品サイトからの安全なアトリビューションパラメータは継続時に保持されます。',
    demo_try_h2: 'できること',
    demo_try_1: 'ブランド付きAIアバターエージェントを選択',
    demo_try_2: 'テキストで製品スタイルの質問',
    demo_try_3: '有効な場合は音声インタラクション',
    demo_try_4: '訪問者向けインバウンドAI対応の感覚を体験',
    demo_after_h2: 'デモの後',
    demo_after_1: '営業とガイド付きウォークスルーを予約',
    demo_after_2: 'テナントワークスペースを登録',
    demo_after_3: 'ライブカタログからパッケージを選択',
    demo_after_4: 'チェックアウトは認証済み請求のまま',
    demo_open_h2: 'ライブデモを開く',
    demo_open_p: 'マーケティングサイトを離れ、サイトルートの顧客デモ画面に移動します。',
    demo_launch: 'ライブデモを起動',
    demo_book: 'ガイド付きデモを予約',
    demo_register: '登録を開始'
  }
};
