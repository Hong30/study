CREATE TABLE `timeline` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `weibo_user_id` int(11) NOT NULL,
  `weibo_id` int(11) NOT NULL,
  `weibo_created_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;
