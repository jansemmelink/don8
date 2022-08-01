CREATE DATABASE IF NOT EXISTS `don8`;
GRANT ALL PRIVILEGES ON `don8`.* to 'don8'@'%' IDENTIFIED BY 'don8';

DROP TABLE IF EXISTS `groups`;
CREATE TABLE `groups` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `parent_group_id` VARCHAR(40) DEFAULT NULL,
  `parent_title` VARCHAR(255) DEFAULT NULL,
  `title` VARCHAR(100) NOT NULL,
  `description` VARCHAR(255) DEFAULT NULL,
  `search` VARCHAR(255) NOT NULL,
  UNIQUE KEY `group_id` (`id`),
  UNIQUE KEY `group_title` (`parent_title`,`title`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `phone` VARCHAR(15) NOT NULL,
  UNIQUE KEY `user_id` (`id`),
  UNIQUE KEY `user_phone` (`phone`)  
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

INSERT INTO `users` SET `name`="admin",`phone`="0824526299";

DROP TABLE IF EXISTS `coordinators`;
CREATE TABLE `coordinators` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `group_id` VARCHAR(40) NOT NULL,
  `user_id` VARCHAR(40) NOT NULL,
  `role` VARCHAR(100) NOT NULL,
  UNIQUE KEY `coordinator_id` (`id`),
  UNIQUE KEY `coordinator_group_user` (`group_id`,`user_id`),
  FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`),
  FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

DROP TABLE IF EXISTS `coordinator_permissions`;
CREATE TABLE `coordinator_permissions` (
  `coordinator_id` VARCHAR(40) NOT NULL,
  `permissions` VARCHAR(100) NOT NULL,
  UNIQUE KEY `coordinator_permission` (`coordinator_id`,`permissions`),
  FOREIGN KEY (`coordinator_id`) REFERENCES `coordinators`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

DROP TABLE IF EXISTS `locations`;
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

DROP TABLE IF EXISTS `location_schedules`;
CREATE TABLE `location_schedules` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `location_id` VARCHAR(40) NOT NULL,
  `open_time` DATETIME NOT NULL,
  `close_time` DATETIME NOT NULL,
  `coordinator_id` VARCHAR(40) NOT NULL,
  UNIQUE KEY `location_schedule_id` (`id`),
  FOREIGN KEY (`location_id`) REFERENCES `locations`(`id`),
  FOREIGN KEY (`coordinator_id`) REFERENCES `coordinators`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

DROP TABLE IF EXISTS `requests`;
CREATE TABLE `requests` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `group_id` VARCHAR(40) NOT NULL,
  `title` VARCHAR(100) NOT NULL,
  `description` VARCHAR(255) DEFAULT NULL,
  `tags` VARCHAR(255) DEFAULT NULL,
  `unit` VARCHAR(100) DEFAULT NULL,
  `qty` INT(11) DEFAULT 0,
  UNIQUE KEY `request_id` (`id`),
  UNIQUE KEY `request_title` (`group_id`,`title`),
  FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

DROP TABLE IF EXISTS `promises`;
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

DROP TABLE IF EXISTS `receives`;
CREATE TABLE `receives` (
  `id` VARCHAR(40) DEFAULT (uuid()) NOT NULL,
  `location_id` VARCHAR(40) NOT NULL,
  `request_id` VARCHAR(40) DEFAULT NULL,
  `promise_id` VARCHAR(40) DEFAULT NULL,
  `title` VARCHAR(100) NOT NULL,
  `unit` VARCHAR(100) NOT NULL,
  `qty` INT(11) NOT NULL,
  UNIQUE KEY `receive_id` (`id`),
  FOREIGN KEY (`location_id`) REFERENCES `locations`(`id`),
  FOREIGN KEY (`request_id`) REFERENCES `requests`(`id`),
  FOREIGN KEY (`promise_id`) REFERENCES `promises`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;

DROP TABLE IF EXISTS `logs`;
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

