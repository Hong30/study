CREATE TABLE `collect` (
  `user_id` int(11) NOT NULL,
  `weibo_id` int(11) NOT NULL,
  `created_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`user_id`,`weibo_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;