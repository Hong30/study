CREATE TABLE `users` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `account` varchar(16) NOT NULL,
  `avatar` varchar(255) NOT NULL,
  `password` varchar(64) NOT NULL,
  `salt` varchar(16) NOT NULL,
  `following_num` int(11) unsigned NOT NULL DEFAULT '0',
  `follower_num` int(11) unsigned NOT NULL DEFAULT '0',
  `weibo_num` int(11) unsigned NOT NULL DEFAULT '0',
  `created_at` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;