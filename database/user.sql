CREATE TABLE `weibo`.`users` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `account` VARCHAR(16) NOT NULL,
  `avatar` VARCHAR(255) NOT NULL,
  `password` VARCHAR(64) NOT NULL,
  `salt` VARCHAR(16) NOT NULL,
  `created_at` INT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`));