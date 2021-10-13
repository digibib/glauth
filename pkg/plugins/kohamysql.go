package main

import (
  "database/sql"
  "fmt"

  _ "github.com/go-sql-driver/mysql"

  "github.com/glauth/glauth/pkg/handler"
)

// TODO: categories -> groups
type KohaMysqlBackend struct {
}

func NewKohaMySQLHandler(opts ...handler.Option) handler.Handler {
  backend := KohaMysqlBackend{}
  return NewDatabaseHandler(backend, opts...)
}

func (b KohaMysqlBackend) GetDriverName() string {
  return "mysql"
}

// Lookup user by cardnumber
func (b KohaMysqlBackend) FindUserQuery(criterion string) string {
  // uidnumber, primarygroup, passbcrypt, passsha256, otpsecret, yubikey
  return fmt.Sprintf("SELECT borrowernumber,6001,password,'','','' FROM borrowers WHERE password IS NOT NULL AND lower(cardnumber)=?")
}

func (b KohaMysqlBackend) FindGroupQuery() string {
  return `SELECT CASE ?
    WHEN 'admins' THEN 6000
    WHEN 'borrowers' THEN 6001
    ELSE 6002
    END`
}

// Fetch all members, used in Sync
func (b KohaMysqlBackend) FindPosixAccountsQuery() string {
  // name, uidnumber, primarygroup, passbcrypt, passsha256, otpsecret, yubikey, othergroups, givenname, sn, mail, loginshell, homedirectory, disabled
  return `SELECT cardnumber,borrowernumber,6001,'','','','','',
    IFNULL(firstname,''),IFNULL(surname,''),IFNULL(alertemail,''),'','',0
    FROM borrowers
    WHERE deleted_at IS NULL
    AND categorycode = 'V'
    AND password IS NOT NULL`
}

// Fetch all groups
func (b KohaMysqlBackend) MemoizeGroupsQuery() string {
  return `SELECT * FROM ( VALUES
    ('admins',6000,6001),
    ('admins',6000,6002),
    ('borrowers',6001,6002),
    ('others',6002,NULL)
  ) t`
}

// Fetch User by cardnumber
func (b KohaMysqlBackend) GetGroupMembersQuery() string {
  //  name, uidnumber, primarygroup, passbcrypt, passsha256, otpsecret, yubikey, othergroups
  return "SELECT cardnumber,borrowernumber,6001,password,'','','','' FROM borrowers WHERE password IS NOT NULL AND lower(cardnumber)=?"
}

// Fetch all user IDs, basically a short version of FindPosixAccountsQuery
func (b KohaMysqlBackend) GetGroupMemberIDsQuery() string {
  // name, uidnumber, primarygroup, passbcrypt, passsha256, otpsecret, yubikey, othergroups
  return `SELECT cardnumber,borrowernumber,6001,'','','','',''
  FROM borrowers
  WHERE deleted_at IS NULL
  AND categorycode = 'V'
  AND password IS NOT NULL`
}

// Get capabilities of user
// TODO: make some decisions on proper access levels
func (b KohaMysqlBackend) GetUserCapabilitiesQuery() string {
  // action,object
  return "SELECT * FROM ( VALUES ('search','ou=borrowers,dc=obib,dc=no'),('search','*') ) t WHERE ?"
}

// Create db/schema if necessary
func (b KohaMysqlBackend) CreateSchema(db *sql.DB) {
  statement, _ := db.Prepare(
    `CREATE TABLE IF NOT EXISTS borrowers (
  borrowernumber int(11) NOT NULL auto_increment,
  cardnumber varchar(32) default NULL,
  surname mediumtext NOT NULL,
  firstname text,
  email mediumtext,
  dateofbirth date default NULL,
  branchcode varchar(10) NOT NULL default '',
  categorycode varchar(10) NOT NULL default '',
  password varchar(60) default NULL,
  smsalertnumber varchar(50) default NULL,
  deleted_at datetime DEFAULT NULL,

  PRIMARY KEY borrowernumber (borrowernumber),
  UNIQUE KEY cardnumber (cardnumber),
  KEY categorycode (categorycode),
  KEY branchcode (branchcode),
  KEY borr_email (email),
  KEY borr_sms (smsalertnumber),
  KEY surname_idx (surname(255)),
  KEY firstname_idx (firstname(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`)
  statement.Exec()
}
