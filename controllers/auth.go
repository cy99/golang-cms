package controllers

import (
	"fmt"

	"github.com/dionyself/golang-cms/models"
	"github.com/dionyself/golang-cms/utils"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/orm"
)

var sessionName = beego.AppConfig.String("SessionName")

type LoginController struct {
	BaseController
}

func (CTRL *LoginController) LoginView() {
	CTRL.ConfigPage("login.html")

}

func (CTRL *LoginController) Login() {
	username := CTRL.GetString("username")
	password := CTRL.GetString("password")
	backTo := CTRL.GetString("back_to")

	var user models.User
	if VerifyUser(&user, username, password) {
		CTRL.SetSession(sessionName, user.Id)
		if backTo != "" {
			CTRL.Redirect("/"+backTo, 302)
		} else {
			CTRL.Redirect("/profile/0/show", 302)
		}
	} else {
		CTRL.Redirect("/register", 302)
	}

}

func (CTRL *LoginController) Logout() {
	CTRL.DelSession(sessionName)
	CTRL.Redirect("/login", 302)
}

func (CTRL *LoginController) RegisterView() {
	CTRL.ConfigPage("register.html")
}

func (CTRL *LoginController) Register() {
	form := models.RegisterForm{}
	if err := CTRL.ParseForm(&form); err != nil {
		CTRL.Abort("401")
	}

	if form.Validate() {
		salt := utils.GetRandomString(10)
		encodedPwd := salt + "$" + utils.EncodePassword(form.Password, salt)

		o := CTRL.GetDB()
		profile := new(models.Profile)
		profile.Age = 0
		profile.Avatar = "female"
		if form.Gender {
			profile.Avatar = "male"
		}
		profile.Gender = form.Gender
		user := new(models.User)
		user.Profile = profile
		user.Username = form.Username
		user.Password = encodedPwd
		user.Rands = salt
		fmt.Println(o.Insert(profile))
		fmt.Println(o.Insert(user))

		CTRL.Redirect("/", 302)

	} else {
		CTRL.Data["form"] = form
		for key, msg := range form.InvalidFields {
			fmt.Println(key, msg)
		}
		CTRL.ConfigPage("register.html")
	}
}

func HasUser(user *models.User, username string) bool {
	var err error
	qs := orm.NewOrm()
	user.Username = username
	err = qs.Read(user, "Username")
	if err == nil {
		return true
	}
	return false
}

func VerifyPassword(rawPwd, encodedPwd string) bool {
	var salt, encoded string
	salt = encodedPwd[:10]
	encoded = encodedPwd[11:]
	return utils.EncodePassword(rawPwd, salt) == encoded
}

func VerifyUser(user *models.User, username, password string) (success bool) {
	// search user by username or email
	if HasUser(user, username) == false {
		return
	}
	if VerifyPassword(password, user.Password) {
		success = true
	}
	return
}

var AuthRequest = func(ctx *context.Context) {
	uid, ok := ctx.Input.Session(sessionName).(int)
	if !ok && ctx.Input.URI() != "/login" && ctx.Input.URI() != "/register" {
		ctx.Redirect(302, "/login")
		return
	}
	var user models.User
	var err error
	qs := orm.NewOrm()
	user.Id = uid
	err = qs.Read(&user, "Id")
	if err != nil {
		ctx.Redirect(302, "/login")
		return
	}
	qs.LoadRelated(&user, "Profile")
	ctx.Input.SetData("user", user)
}
