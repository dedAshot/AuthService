# AuthService

It is a test task

Task link: https://medods.notion.site/Test-task-BackDev-623508ed85474f48a721e43ab00e9916

To run the test you need db with installed volume scripts/authservice_pginit/create_table_users.sql 
Test db configuration insede the tests

The service supplies 2 paths:
1) GET [host]/gettoken/?GUID=[your user guid] 
  returning token pair { "access": "your token", "refresh": "your token"} if the guid exists
3) POST [host]/refreshtoken/  body: { "access": "your token", "refresh": "your token"}
  returning token pair { "access": "your token", "refresh": "your token"} if the first pair valid
