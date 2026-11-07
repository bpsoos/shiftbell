package profiles

func AvatarHref(id int) string {
	avatars := [...]string{"gold", "plum", "sage", "teal", "terracotta"}
	return "/assets/avatars/" + avatars[(id-1)%len(avatars)] + ".svg"
}
