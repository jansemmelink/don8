CREATE DATABASE IF NOT EXISTS `don8`;
GRANT ALL PRIVILEGES ON `don8`.* to 'don8'@'%' IDENTIFIED BY 'don8';

DROP TABLE IF EXISTS `logs`;
DROP TABLE IF EXISTS `receives`;
DROP TABLE IF EXISTS `promises`;
DROP TABLE IF EXISTS `requests`;
DROP TABLE IF EXISTS `location_schedules`;
DROP TABLE IF EXISTS `locations`;
DROP TABLE IF EXISTS `member_permissions`;
DROP TABLE IF EXISTS `members`;
DROP TABLE IF EXISTS `invitations`;
DROP TABLE IF EXISTS `groups`;
DROP TABLE IF EXISTS `sessions`;
DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `phone` VARCHAR(15) NOT NULL,
  `email` VARCHAR(100) NOT NULL,
  `tpw` VARCHAR(40) DEFAULT NULL,
  `tpw_exp` DATETIME DEFAULT NULL,
  `pwd_hash` VARCHAR(40) DEFAULT NULL,
  UNIQUE KEY `user_id` (`id`),
  UNIQUE KEY `user_phone` (`phone`),
  UNIQUE KEY `user_email` (`email`),
  UNIQUE KEY `user_tpw` (`tpw`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `sessions` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `user_id` VARCHAR(40) DEFAULT NULL,
  `start_time` DATETIME NOT NULL,
  `expiry_time` DATETIME NOT NULL,
  UNIQUE KEY `session_id` (`id`),
  UNIQUE KEY `session_user` (`user_id`),
  FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `groups` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `parent_group_id` VARCHAR(40) DEFAULT NULL,
  `title` VARCHAR(100) NOT NULL,
  `description` VARCHAR(255) DEFAULT NULL,
  `start` DATETIME DEFAULT NULL,
  `end` DATETIME DEFAULT NULL,
  UNIQUE KEY `group_id` (`id`),
  UNIQUE KEY `group_title` (`parent_group_id`,`title`),
  KEY `group_start` (`start`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `invitations` (
  `id` VARCHAR(40) NOT NULL,
  `group_id` VARCHAR(40) NOT NULL,
  `email` VARCHAR(100) NOT NULL,
  `time_created` DATETIME NOT NULL,
  `time_updated` DATETIME NOT NULL,
  `status` VARCHAR(30) NOT NULL,
  UNIQUE KEY `invitation_id` (`id`),
  UNIQUE KEY `invitation_uniq` (`group_id`,`email`),
  FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `members` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `group_id` VARCHAR(40) NOT NULL,
  `user_id` VARCHAR(40) NOT NULL,
  `role` VARCHAR(100) NOT NULL,
  UNIQUE KEY `member_id` (`id`),
  UNIQUE KEY `member_group_user` (`group_id`,`user_id`),
  FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`),
  FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `member_permissions` (
  `member_id` VARCHAR(40) NOT NULL,
  `permissions` VARCHAR(100) NOT NULL,
  UNIQUE KEY `member_permission` (`member_id`,`permissions`),
  FOREIGN KEY (`member_id`) REFERENCES `members`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `locations` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `group_id` VARCHAR(40) NOT NULL,
  `title` VARCHAR(100) NOT NULL,
  `description` VARCHAR(255) DEFAULT NULL,
  `final_destination` TINYINT(1) DEFAULT 0,
  UNIQUE KEY `location_id` (`id`),
  UNIQUE KEY `group_location` (`group_id`,`title`),
  FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `location_schedules` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `location_id` VARCHAR(40) NOT NULL,
  `open_time` DATETIME NOT NULL,
  `close_time` DATETIME NOT NULL,
  `member_id` VARCHAR(40) NOT NULL,
  UNIQUE KEY `location_schedule_id` (`id`),
  FOREIGN KEY (`location_id`) REFERENCES `locations`(`id`),
  FOREIGN KEY (`member_id`) REFERENCES `members`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `requests` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `group_id` VARCHAR(40) NOT NULL,
  `title` VARCHAR(100) NOT NULL,
  `description` VARCHAR(255) DEFAULT NULL,
  `tags` VARCHAR(255) DEFAULT NULL,
  `units` VARCHAR(100) DEFAULT NULL,
  `qty` INT(11) DEFAULT 0,
  UNIQUE KEY `request_id` (`id`),
  UNIQUE KEY `request_title` (`group_id`,`title`),
  FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `promises` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `request_id` VARCHAR(40) NOT NULL,
  `user_id` VARCHAR(40) NOT NULL,
  `location_id` VARCHAR(40) DEFAULT NULL,
  `qty` INT(11) NOT NULL,
  `date` DATETIME NOT NULL,
  UNIQUE KEY `promise_id` (`id`),
  FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
  FOREIGN KEY (`request_id`) REFERENCES `requests`(`id`),
  FOREIGN KEY (`location_id`) REFERENCES `locations`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `receives` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `location_id` VARCHAR(40) NOT NULL,
  `request_id` VARCHAR(40) DEFAULT NULL,
  `promise_id` VARCHAR(40) DEFAULT NULL,
  `title` VARCHAR(100) NOT NULL,
  `qty` INT(11) NOT NULL,
  UNIQUE KEY `receive_id` (`id`),
  FOREIGN KEY (`location_id`) REFERENCES `locations`(`id`),
  FOREIGN KEY (`request_id`) REFERENCES `requests`(`id`),
  FOREIGN KEY (`promise_id`) REFERENCES `promises`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

CREATE TABLE `logs` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `table` VARCHAR(100) NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `user_id` VARCHAR(40) NOT NULL,
  `action` VARCHAR(100) NOT NULL,
  `values` JSON DEFAULT NULL,
  UNIQUE KEY `log_id` (`id`),
  UNIQUE KEY `log_index` (`id`,`table`,`timestamp`),
  FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

