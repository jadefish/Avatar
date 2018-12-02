CREATE TABLE IF NOT EXISTS `accounts` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'UTC timestamp of record creation',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'UTC timestamp of last update',
    `deleted_at` timestamp NULL DEFAULT NULL COMMENT 'UTC timestamp of soft record deletion',
    `last_logged_in_at` timestamp NULL DEFAULT NULL COMMENT 'UTC timestamp of last successful login',
    `name` varchar(20) CHARACTER SET ascii NOT NULL COMMENT 'Account name, limited to 20 characters',
    `email` varchar(254) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Account owner''s registered email address',
    `password_hash` varchar(60) CHARACTER SET ascii NOT NULL COMMENT 'Password hash',
    `creation_ip` varbinary(16) NOT NULL COMMENT 'IP address of record creation',
    `last_login_ip` varbinary(16) DEFAULT NULL COMMENT 'IP address of last successful login',
    PRIMARY KEY (`id`),
    UNIQUE KEY `name` (`name`),
    UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
