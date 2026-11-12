// Gophish i18n - Internationalization Support
// Supports: English (en), Chinese (zh)

const i18n = {
    currentLang: localStorage.getItem('gophish_lang') || 'zh',

    translations: {
        en: {
            // Navigation
            nav_dashboard: 'Dashboard',
            nav_campaigns: 'Campaigns',
            nav_users: 'Users & Groups',
            nav_templates: 'Email Templates',
            nav_landing_pages: 'Landing Pages',
            nav_sending_profiles: 'Sending Profiles',
            nav_webhooks: 'Webhooks',
            nav_settings: 'Account Settings',
            nav_user_management: 'User Management',
            nav_logout: 'Logout',
            nav_language: 'Language',

            // Common
            btn_save: 'Save',
            btn_cancel: 'Cancel',
            btn_close: 'Close',
            btn_delete: 'Delete',
            btn_edit: 'Edit',
            btn_copy: 'Copy',
            btn_new: 'New',
            btn_import: 'Import',
            btn_export: 'Export',
            btn_test: 'Test',
            btn_launch: 'Launch',
            btn_send: 'Send',
            btn_ok: 'OK',
            btn_confirm: 'Confirm',

            loading: 'Loading...',
            no_data: 'No data available',
            success: 'Success',
            error: 'Error',
            warning: 'Warning',

            // Dashboard
            dashboard_title: 'Dashboard',
            dashboard_campaigns: 'Recent Campaigns',
            dashboard_total_campaigns: 'Total Campaigns',
            dashboard_emails_sent: 'Emails Sent',
            dashboard_emails_opened: 'Emails Opened',
            dashboard_links_clicked: 'Links Clicked',
            dashboard_data_submitted: 'Data Submitted',
            dashboard_no_campaigns: 'No campaigns created yet',

            // Campaigns
            campaigns_title: 'Campaigns',
            campaigns_new: 'New Campaign',
            campaigns_batch_delete: 'Batch Delete',
            campaigns_active: 'Active Campaigns',
            campaigns_archived: 'Archived Campaigns',
            campaigns_name: 'Name',
            campaigns_created_date: 'Created Date',
            campaigns_status: 'Status',
            campaigns_email_template: 'Email Template',
            campaigns_landing_page: 'Landing Page',
            campaigns_url: 'URL',
            campaigns_url_tip: 'Location of Gophish listener (must be reachable by targets!)',
            campaigns_qr_size: 'QR Code Size',
            campaigns_qr_size_tip: 'Size of QR code images (integer is height & width in pixels). Leave blank to not include QR code images and use normal links.',
            campaigns_launch_date: 'Launch Date',
            campaigns_send_by: 'Send Emails By (Optional)',
            campaigns_send_by_tip: 'If specified, Gophish will send emails evenly between the campaign launch and this date.',
            campaigns_sending_profile: 'Sending Profile',
            campaigns_groups: 'Groups',
            campaigns_select_template: 'Select a Template',
            campaigns_select_page: 'Select a Landing Page',
            campaigns_select_profile: 'Select a Sending Profile',
            campaigns_select_groups: 'Select Groups',
            campaigns_launch_confirm: 'Are you sure?',
            campaigns_launch_confirm_text: 'This will schedule the campaign to be launched.',
            campaigns_launch_success: 'Campaign Scheduled!',
            campaigns_launch_success_text: 'This campaign has been scheduled for launch!',
            campaigns_delete_confirm: 'Are you sure?',
            campaigns_delete_confirm_text: "This will delete the campaign. This can't be undone!",
            campaigns_delete_success: 'Campaign Deleted!',
            campaigns_delete_success_text: 'This campaign has been deleted!',
            campaigns_no_active: 'No campaigns created yet. Let\'s create one!',
            campaigns_no_archived: 'No archived campaigns.',
            campaigns_send_test: 'Send Test Email',
            campaigns_test_email_to: 'Send Test Email to:',
            campaigns_test_first_name: 'First Name',
            campaigns_test_last_name: 'Last Name',
            campaigns_test_email: 'Email',
            campaigns_test_position: 'Position',
            campaigns_test_sent: 'Email Sent!',
            campaigns_view_results: 'View Results',
            campaigns_copy: 'Copy Campaign',
            campaigns_status_in_progress: 'In progress',
            campaigns_status_queued: 'Queued',
            campaigns_status_completed: 'Completed',
            campaigns_status_sent: 'Emails Sent',
            campaigns_status_error: 'Error',

            // Campaign Results
            results_title: 'Campaign Results',
            results_overview: 'Overview',
            results_details: 'Details',
            results_timeline: 'Timeline',
            results_email: 'Email',
            results_status: 'Status',
            results_ip: 'IP Address',
            results_position: 'Position',
            results_reported: 'Reported',
            results_complete: 'Complete Campaign',
            results_delete: 'Delete Campaign',
            results_export: 'Export CSV',
            results_sent: 'Sent',
            results_opened: 'Opened',
            results_clicked: 'Clicked',
            results_submitted: 'Submitted Data',
            results_reported_event: 'Email Reported',
            results_error: 'Error',

            // Templates
            templates_title: 'Email Templates',
            templates_new: 'New Template',
            templates_name: 'Name',
            templates_modified: 'Last Modified',
            templates_envelope_sender: 'Envelope Sender',
            templates_envelope_sender_tip: 'This sender is shown in email headers',
            templates_subject: 'Subject',
            templates_import_email: 'Import Email',
            templates_add_tracking: 'Add Tracking Image',
            templates_add_files: 'Add Files',
            templates_text: 'Text',
            templates_html: 'HTML',
            templates_source: 'Source',
            templates_preview: 'Preview',
            templates_delete_confirm: 'Delete Template?',
            templates_delete_confirm_text: "This will delete the template. This can't be undone!",
            templates_delete_success: 'Template deleted successfully!',
            templates_save_success: 'Template saved successfully!',
            templates_no_templates: 'No templates created yet. Let\'s create one!',

            // Landing Pages
            pages_title: 'Landing Pages',
            pages_new: 'New Page',
            pages_name: 'Name',
            pages_modified: 'Last Modified',
            pages_import_site: 'Import Site',
            pages_capture_credentials: 'Capture Submitted Data',
            pages_capture_passwords: 'Capture Passwords',
            pages_redirect: 'Redirect to',
            pages_redirect_tip: 'URL to redirect users after submitting credentials',
            pages_source: 'Source',
            pages_preview: 'Preview',
            pages_delete_confirm: 'Delete Landing Page?',
            pages_delete_confirm_text: "This will delete the landing page. This can't be undone!",
            pages_delete_success: 'Landing page deleted successfully!',
            pages_save_success: 'Landing page saved successfully!',
            pages_no_pages: 'No landing pages created yet. Let\'s create one!',

            // Groups
            groups_title: 'Users & Groups',
            groups_new: 'New Group',
            groups_name: 'Group Name',
            groups_targets: 'Number of Members',
            groups_modified: 'Last Modified',
            groups_bulk_import: 'Bulk Import Users',
            groups_add_user: 'Add',
            groups_first_name: 'First Name',
            groups_last_name: 'Last Name',
            groups_email: 'Email',
            groups_position: 'Position',
            groups_delete_confirm: 'Delete Group?',
            groups_delete_confirm_text: "This will delete the group. This can't be undone!",
            groups_delete_success: 'Group deleted successfully!',
            groups_save_success: 'Group saved successfully!',
            groups_no_groups: 'No groups created yet. Let\'s create one!',

            // Sending Profiles
            profiles_title: 'Sending Profiles',
            profiles_new: 'New Profile',
            profiles_name: 'Name',
            profiles_interface_type: 'Interface Type',
            profiles_smtp: 'SMTP',
            profiles_from: 'From',
            profiles_host: 'Host',
            profiles_username: 'Username',
            profiles_password: 'Password',
            profiles_ignore_cert: 'Ignore Certificate Errors',
            profiles_custom_headers: 'Email Headers',
            profiles_add_header: 'Add Custom Header',
            profiles_send_test: 'Send Test Email',
            profiles_delete_confirm: 'Delete Sending Profile?',
            profiles_delete_confirm_text: "This will delete the sending profile. This can't be undone!",
            profiles_delete_success: 'Profile deleted successfully!',
            profiles_save_success: 'Profile saved successfully!',
            profiles_no_profiles: 'No sending profiles created yet. Let\'s create one!',
            profiles_test_success: 'Email Sent!',

            // Settings
            settings_title: 'Settings',
            settings_account: 'Account Settings',
            settings_ui: 'UI Settings',
            settings_reporting: 'Reporting Settings',
            settings_version: 'Gophish version',
            settings_api_key: 'API Key',
            settings_api_reset: 'Reset',
            settings_username: 'Username',
            settings_old_password: 'Old Password',
            settings_new_password: 'New Password',
            settings_confirm_password: 'Confirm New Password',
            settings_show_map: 'Show campaign results map',
            settings_imap: 'Enable Email Account Monitoring',
            settings_imap_host: 'IMAP Host',
            settings_imap_port: 'IMAP Port',
            settings_imap_username: 'IMAP Username',
            settings_imap_password: 'IMAP Password',
            settings_use_tls: 'Use TLS',
            settings_save_success: 'Settings saved successfully!',

            // Users (Admin)
            users_title: 'User Management',
            users_new: 'New User',
            users_username: 'Username',
            users_role: 'Role',
            users_password: 'Password',
            users_confirm_password: 'Confirm Password',
            users_delete_confirm: 'Delete User?',
            users_delete_confirm_text: "This will delete the user. This can't be undone!",
            users_delete_success: 'User deleted successfully!',
            users_save_success: 'User saved successfully!',
            users_no_users: 'No users found.',

            // Webhooks
            webhooks_title: 'Webhooks',
            webhooks_new: 'New Webhook',
            webhooks_name: 'Name',
            webhooks_url: 'URL',
            webhooks_secret: 'Secret',
            webhooks_active: 'Is Active',
            webhooks_delete_confirm: 'Delete Webhook?',
            webhooks_delete_confirm_text: "This will delete the webhook. This can't be undone!",
            webhooks_delete_success: 'Webhook deleted successfully!',
            webhooks_save_success: 'Webhook saved successfully!',
            webhooks_no_webhooks: 'No webhooks created yet. Let\'s create one!',

            // Login
            login_title: 'Login',
            login_username: 'Username',
            login_password: 'Password',
            login_button: 'Login',
            login_error: 'Invalid username or password',

            // Errors
            error_required: 'This field is required',
            error_email_invalid: 'Please enter a valid email address',
            error_no_groups: 'No groups found!',
            error_no_templates: 'No templates found!',
            error_no_pages: 'No pages found!',
            error_no_profiles: 'No profiles found!',
            error_fetching: 'Error fetching data',
        },

        zh: {
            // 导航菜单
            nav_dashboard: '仪表盘',
            nav_campaigns: '钓鱼活动',
            nav_users: '用户和组',
            nav_templates: '邮件模板',
            nav_landing_pages: '钓鱼页面',
            nav_sending_profiles: '发送配置',
            nav_webhooks: '网络钩子',
            nav_settings: '账户设置',
            nav_user_management: '用户管理',
            nav_logout: '退出登录',
            nav_language: '语言',

            // 通用按钮
            btn_save: '保存',
            btn_cancel: '取消',
            btn_close: '关闭',
            btn_delete: '删除',
            btn_edit: '编辑',
            btn_copy: '复制',
            btn_new: '新建',
            btn_import: '导入',
            btn_export: '导出',
            btn_test: '测试',
            btn_launch: '启动',
            btn_send: '发送',
            btn_ok: '确定',
            btn_confirm: '确认',

            loading: '加载中...',
            no_data: '暂无数据',
            success: '成功',
            error: '错误',
            warning: '警告',

            // 仪表盘
            dashboard_title: '仪表盘',
            dashboard_campaigns: '近期活动',
            dashboard_total_campaigns: '活动总数',
            dashboard_emails_sent: '已发送邮件',
            dashboard_emails_opened: '邮件被打开',
            dashboard_links_clicked: '链接点击',
            dashboard_data_submitted: '数据提交',
            dashboard_no_campaigns: '暂无活动',

            // 钓鱼活动
            campaigns_title: '钓鱼活动',
            campaigns_new: '新建活动',
            campaigns_batch_delete: '批量删除',
            campaigns_active: '进行中的活动',
            campaigns_archived: '已归档活动',
            campaigns_name: '名称',
            campaigns_created_date: '创建时间',
            campaigns_status: '状态',
            campaigns_email_template: '邮件模板',
            campaigns_landing_page: '钓鱼页面',
            campaigns_url: '钓鱼链接',
            campaigns_url_tip: 'Gophish 监听地址（目标必须能够访问！）',
            campaigns_qr_size: '二维码大小',
            campaigns_qr_size_tip: '二维码图片大小（像素）。留空则不生成二维码，使用普通链接。',
            campaigns_launch_date: '启动时间',
            campaigns_send_by: '发送截止时间（可选）',
            campaigns_send_by_tip: '如果设置，Gophish 将在启动时间和此时间之间均匀发送邮件。',
            campaigns_sending_profile: '发送配置',
            campaigns_groups: '目标用户组',
            campaigns_select_template: '选择邮件模板',
            campaigns_select_page: '选择钓鱼页面',
            campaigns_select_profile: '选择发送配置',
            campaigns_select_groups: '选择用户组',
            campaigns_launch_confirm: '确认启动？',
            campaigns_launch_confirm_text: '活动将按计划启动并发送钓鱼邮件。',
            campaigns_launch_success: '活动已创建！',
            campaigns_launch_success_text: '钓鱼活动已安排启动！',
            campaigns_delete_confirm: '确认删除？',
            campaigns_delete_confirm_text: '此操作将永久删除该活动，无法恢复！',
            campaigns_delete_success: '活动已删除！',
            campaigns_delete_success_text: '钓鱼活动已成功删除！',
            campaigns_no_active: '暂无活动，点击新建开始第一个活动！',
            campaigns_no_archived: '暂无已归档活动。',
            campaigns_send_test: '发送测试邮件',
            campaigns_test_email_to: '发送测试邮件到：',
            campaigns_test_first_name: '名字',
            campaigns_test_last_name: '姓氏',
            campaigns_test_email: '邮箱地址',
            campaigns_test_position: '职位',
            campaigns_test_sent: '测试邮件已发送！',
            campaigns_view_results: '查看结果',
            campaigns_copy: '复制活动',
            campaigns_status_in_progress: '进行中',
            campaigns_status_queued: '队列中',
            campaigns_status_completed: '已完成',
            campaigns_status_sent: '已发送',
            campaigns_status_error: '错误',

            // 活动结果
            results_title: '活动结果',
            results_overview: '概览',
            results_details: '详情',
            results_timeline: '时间线',
            results_email: '邮箱',
            results_status: '状态',
            results_ip: 'IP 地址',
            results_position: '职位',
            results_reported: '已举报',
            results_complete: '结束活动',
            results_delete: '删除活动',
            results_export: '导出 CSV',
            results_sent: '已发送',
            results_opened: '已打开',
            results_clicked: '已点击',
            results_submitted: '已提交数据',
            results_reported_event: '举报邮件',
            results_error: '错误',

            // 邮件模板
            templates_title: '邮件模板',
            templates_new: '新建模板',
            templates_name: '模板名称',
            templates_modified: '最后修改',
            templates_envelope_sender: '信封发件人',
            templates_envelope_sender_tip: '此发件人显示在邮件头中',
            templates_subject: '邮件主题',
            templates_import_email: '导入邮件',
            templates_add_tracking: '添加追踪图片',
            templates_add_files: '添加附件',
            templates_text: '纯文本',
            templates_html: 'HTML',
            templates_source: '源代码',
            templates_preview: '预览',
            templates_delete_confirm: '确认删除模板？',
            templates_delete_confirm_text: '此操作将永久删除该模板，无法恢复！',
            templates_delete_success: '模板已成功删除！',
            templates_save_success: '模板已成功保存！',
            templates_no_templates: '暂无模板，点击新建开始创建！',

            // 钓鱼页面
            pages_title: '钓鱼页面',
            pages_new: '新建页面',
            pages_name: '页面名称',
            pages_modified: '最后修改',
            pages_import_site: '克隆网站',
            pages_capture_credentials: '捕获提交数据',
            pages_capture_passwords: '捕获密码',
            pages_redirect: '跳转地址',
            pages_redirect_tip: '用户提交凭据后的跳转地址',
            pages_source: '源代码',
            pages_preview: '预览',
            pages_delete_confirm: '确认删除页面？',
            pages_delete_confirm_text: '此操作将永久删除该页面，无法恢复！',
            pages_delete_success: '页面已成功删除！',
            pages_save_success: '页面已成功保存！',
            pages_no_pages: '暂无钓鱼页面，点击新建开始创建！',

            // 用户组
            groups_title: '用户和组',
            groups_new: '新建用户组',
            groups_name: '组名称',
            groups_targets: '成员数量',
            groups_modified: '最后修改',
            groups_bulk_import: '批量导入用户',
            groups_add_user: '添加',
            groups_first_name: '名字',
            groups_last_name: '姓氏',
            groups_email: '邮箱地址',
            groups_position: '职位',
            groups_delete_confirm: '确认删除用户组？',
            groups_delete_confirm_text: '此操作将永久删除该用户组，无法恢复！',
            groups_delete_success: '用户组已成功删除！',
            groups_save_success: '用户组已成功保存！',
            groups_no_groups: '暂无用户组，点击新建开始创建！',

            // 发送配置
            profiles_title: '发送配置',
            profiles_new: '新建配置',
            profiles_name: '配置名称',
            profiles_interface_type: '接口类型',
            profiles_smtp: 'SMTP',
            profiles_from: '发件人',
            profiles_host: 'SMTP 服务器',
            profiles_username: '用户名',
            profiles_password: '密码',
            profiles_ignore_cert: '忽略证书错误',
            profiles_custom_headers: '自定义邮件头',
            profiles_add_header: '添加邮件头',
            profiles_send_test: '发送测试邮件',
            profiles_delete_confirm: '确认删除发送配置？',
            profiles_delete_confirm_text: '此操作将永久删除该配置，无法恢复！',
            profiles_delete_success: '配置已成功删除！',
            profiles_save_success: '配置已成功保存！',
            profiles_no_profiles: '暂无发送配置，点击新建开始创建！',
            profiles_test_success: '测试邮件已发送！',

            // 设置
            settings_title: '设置',
            settings_account: '账户设置',
            settings_ui: '界面设置',
            settings_reporting: '报告设置',
            settings_version: 'Gophish 版本',
            settings_api_key: 'API 密钥',
            settings_api_reset: '重置',
            settings_username: '用户名',
            settings_old_password: '旧密码',
            settings_new_password: '新密码',
            settings_confirm_password: '确认新密码',
            settings_show_map: '显示活动结果地图',
            settings_imap: '启用邮件监控',
            settings_imap_host: 'IMAP 服务器',
            settings_imap_port: 'IMAP 端口',
            settings_imap_username: 'IMAP 用户名',
            settings_imap_password: 'IMAP 密码',
            settings_use_tls: '使用 TLS',
            settings_save_success: '设置已成功保存！',

            // 用户管理（管理员）
            users_title: '用户管理',
            users_new: '新建用户',
            users_username: '用户名',
            users_role: '角色',
            users_password: '密码',
            users_confirm_password: '确认密码',
            users_delete_confirm: '确认删除用户？',
            users_delete_confirm_text: '此操作将永久删除该用户，无法恢复！',
            users_delete_success: '用户已成功删除！',
            users_save_success: '用户已成功保存！',
            users_no_users: '暂无用户。',

            // 网络钩子
            webhooks_title: '网络钩子',
            webhooks_new: '新建 Webhook',
            webhooks_name: '名称',
            webhooks_url: 'URL 地址',
            webhooks_secret: '密钥',
            webhooks_active: '是否启用',
            webhooks_delete_confirm: '确认删除 Webhook？',
            webhooks_delete_confirm_text: '此操作将永久删除该 Webhook，无法恢复！',
            webhooks_delete_success: 'Webhook 已成功删除！',
            webhooks_save_success: 'Webhook 已成功保存！',
            webhooks_no_webhooks: '暂无 Webhook，点击新建开始创建！',

            // 登录
            login_title: '登录',
            login_username: '用户名',
            login_password: '密码',
            login_button: '登录',
            login_error: '用户名或密码错误',

            // 错误信息
            error_required: '此字段为必填项',
            error_email_invalid: '请输入有效的邮箱地址',
            error_no_groups: '未找到用户组！',
            error_no_templates: '未找到邮件模板！',
            error_no_pages: '未找到钓鱼页面！',
            error_no_profiles: '未找到发送配置！',
            error_fetching: '获取数据失败',
        }
    },

    // Get translation
    t: function (key) {
        const lang = this.currentLang;
        if (this.translations[lang] && this.translations[lang][key]) {
            return this.translations[lang][key];
        }
        // Fallback to English
        if (this.translations.en && this.translations.en[key]) {
            return this.translations.en[key];
        }
        return key;
    },

    // Set language
    setLang: function (lang) {
        if (this.translations[lang]) {
            this.currentLang = lang;
            localStorage.setItem('gophish_lang', lang);
            this.applyTranslations();
            return true;
        }
        return false;
    },

    // Toggle language
    toggleLang: function () {
        const newLang = this.currentLang === 'zh' ? 'en' : 'zh';
        this.setLang(newLang);
    },

    // Apply translations to elements with data-i18n attribute
    applyTranslations: function () {
        document.querySelectorAll('[data-i18n]').forEach(el => {
            const key = el.getAttribute('data-i18n');
            const translation = this.t(key);

            if (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA') {
                if (el.hasAttribute('placeholder')) {
                    el.placeholder = translation;
                } else {
                    el.value = translation;
                }
            } else if (el.tagName === 'OPTION') {
                el.text = translation;
            } else {
                el.textContent = translation;
            }
        });

        // Update data-i18n-title for tooltips
        document.querySelectorAll('[data-i18n-title]').forEach(el => {
            const key = el.getAttribute('data-i18n-title');
            el.setAttribute('title', this.t(key));
        });

        // Update language toggle button text
        const langBtn = document.getElementById('langToggle');
        if (langBtn) {
            langBtn.textContent = this.currentLang === 'zh' ? 'English' : '中文';
        }
    },

    // Initialize
    init: function () {
        document.addEventListener('DOMContentLoaded', () => {
            this.applyTranslations();
        });
    }
};

// Auto-initialize
i18n.init();

// Make it globally available
window.i18n = i18n;
window.t = function (key) { return i18n.t(key); };
