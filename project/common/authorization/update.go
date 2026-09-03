package authorizationpg

//type UpdateArgument struct {
//	Email    string `json:"email,omitempty" validate:"email,max=128"`
//	Password string `json:"password,omitempty" validate:"min=6,max=128"`
//}
//
//func (a UpdateArgument) Validate() error {
//
//	if err := validator.New().Struct(a); err != nil {
//		for _, e := range err.(validator.ValidationErrors) {
//			return errors.New(fmt.Sprintf("Field '%s' does not match validation: %s\n", e.Field(), e.ActualTag()))
//		}
//	}
//
//	return nil
//}
