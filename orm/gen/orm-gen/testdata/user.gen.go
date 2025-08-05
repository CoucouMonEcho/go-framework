package testdata

//
//import (
//	"go-framework/orm"
//	"database/sql"
//)
//
//const (
//	UserName     = "Name"
//	UserAge      = "Age"
//	UserNickName = "NickName"
//	UserPicture  = "Picture"
//)
//
//func UserNameLt(val string) orm.Predicate {
//	return orm.C(UserName).Lt(val)
//}
//
//func UserNameGt(val string) orm.Predicate {
//	return orm.C(UserName).Gt(val)
//}
//
//func UserNameEq(val string) orm.Predicate {
//	return orm.C(UserName).Eq(val)
//}
//
//func UserAgeLt(val *int) orm.Predicate {
//	return orm.C(UserAge).Lt(val)
//}
//
//func UserAgeGt(val *int) orm.Predicate {
//	return orm.C(UserAge).Gt(val)
//}
//
//func UserAgeEq(val *int) orm.Predicate {
//	return orm.C(UserAge).Eq(val)
//}
//
//func UserNickNameLt(val *sql.NullString) orm.Predicate {
//	return orm.C(UserNickName).Lt(val)
//}
//
//func UserNickNameGt(val *sql.NullString) orm.Predicate {
//	return orm.C(UserNickName).Gt(val)
//}
//
//func UserNickNameEq(val *sql.NullString) orm.Predicate {
//	return orm.C(UserNickName).Eq(val)
//}
//
//func UserPictureLt(val []byte) orm.Predicate {
//	return orm.C(UserPicture).Lt(val)
//}
//
//func UserPictureGt(val []byte) orm.Predicate {
//	return orm.C(UserPicture).Gt(val)
//}
//
//func UserPictureEq(val []byte) orm.Predicate {
//	return orm.C(UserPicture).Eq(val)
//}
//
//const (
//	UserDetailAddress = "Address"
//)
//
//func UserDetailAddressLt(val string) orm.Predicate {
//	return orm.C(UserDetailAddress).Lt(val)
//}
//
//func UserDetailAddressGt(val string) orm.Predicate {
//	return orm.C(UserDetailAddress).Gt(val)
//}
//
//func UserDetailAddressEq(val string) orm.Predicate {
//	return orm.C(UserDetailAddress).Eq(val)
//}
