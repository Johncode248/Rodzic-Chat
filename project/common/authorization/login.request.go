package authorizationpg

type LoginRequest struct {
	Login      string `json:"login,omitempty"`
	Password   string `json:"password,omitempty"`
	DeviceId   string `json:"device_id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
}
