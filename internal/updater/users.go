package updater

type UserList struct {
	users []*User
}

type User struct {
	Name        string
	SlackId     string
	SlackName   string
	PagerdutyId string
	Email       string
}

func (u *UserList) ByPagerdutyId(id string) *User {
	for _, user := range u.users {
		if user.PagerdutyId == id {
			return user
		}
	}
	return nil
}
