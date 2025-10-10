package main

func init() {

	var EnglishLits = map[string]string{
		"not_found_title":                "Page not found",
		"not_found_content":              "The page %s could not be found",
		"return_to_home":                 "Return to home",
		"login_page_title":               "Log In",
		"register_page_title":            "Sign in",
		"login_title":                    "Welcome",
		"user":                           "User",
		"password":                       "Password",
		"username_placeholder":           "Enter your username",
		"password_placeholder":           "Enter your password",
		"login":                          "Log In",
		"forgot_password":                "Forgot your password?",
		"sign_up":                        "Sign in",
		"email":                          "Email",
		"confirm_password":               "Confirm your password",
		"confirm_password_placeholder":   "Enter your password again",
		"email_placeholder":              "Enter your email",
		"register_button":                "Sign in",
		"already_have_account":           "I already have an account",
		"home_title":                     "Welcome to Finpilot",
		"title":                          "Finpilot",
		"error_register_empty":           "Fields cannot be empty",
		"user_or_email":                  "User or email address",
		"user_or_email_placeholder":      "Enter your username or email address",
		"error_creating_user":            "Error creating user",
		"invalid_request":                "Invalid request",
		"username_used":                  "Username is already used",
		"email_used":                     "Email is already used",
		"passwords_dont_match":           "The passwords do not match",
		"password_invalid_format":        "Password format is invalid",
		"email_invalid_format":           "Email format is invalid",
		"identifier_or_password_invalid": "El usuario o contraseña no son correctos",
		"log_out":                        "Log out",
		"my_profile":                     "My Profile",
		"online_users":                   "Online Users",
	}

	RegisterLang("en", EnglishLits)
}
