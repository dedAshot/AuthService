SELECT reftokenhash AS hash
FROM users
WHERE users.guid = $1;