CREATE TABLE `shards` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'UTC timestamp of record creation',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'UTC timestamp of last update',
    `deleted_at` timestamp NULL DEFAULT NULL COMMENT 'UTC timestamp of soft record deletion',
    `name` varchar(32) CHARACTER SET ascii NOT NULL DEFAULT '' COMMENT 'Name of the shard, limited to 32 US ASCII characters',
    `time_zone` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Etc/UTC' COMMENT 'Name of the shard''s time time zone (tzdata)',
    `capacity` int(11) unsigned NOT NULL DEFAULT 3000 COMMENT 'Maximum number of players allowed to concurrently exist in the game world',
    `ip_address` varbinary(16) NOT NULL COMMENT 'The IP address of the shard''s game server',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;