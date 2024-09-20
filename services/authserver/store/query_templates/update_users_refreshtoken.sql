UPDATE users 
SET reftokenhash = $1
WHERE users.guid = $2;