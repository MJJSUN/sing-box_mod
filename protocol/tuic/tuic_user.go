package tuic

import (
	"net"

	"github.com/gofrs/uuid/v5"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
)

func (h *Inbound) AddUsers(users []option.TUICUser, ids []int) error {
	for i, user := range users {
		h.userNameList = append(h.userNameList, user.Name)
		h.userPasswordList = append(h.userPasswordList, user.Password) // Save password
		h.uuidToUid[user.UUID] = ids[i]
		h.uidToUuid[ids[i]] = user.UUID
	}
	var userUUIDList [][16]byte
	indexs := make([]int, len(h.userNameList))
	for i, UUID := range h.userNameList {
		indexs[i] = h.uuidToUid[UUID]
		userUUID, err := uuid.FromString(UUID)
		if err != nil {
			return E.Cause(err, "invalid uuid for user ", i)
		}
		userUUIDList = append(userUUIDList, userUUID)
	}
	h.server.UpdateUsers(indexs, userUUIDList, h.userPasswordList) // Use password list instead of name list
	return nil
}

func (h *Inbound) DelUsers(names []string) error {
	if len(names) == 0 {
		return nil
	}
	toDelete := make(map[string]struct{})
	for _, name := range names {
		toDelete[name] = struct{}{}
		delete(h.uidToUuid, h.uuidToUid[name])
		delete(h.uuidToUid, name)
		h.userconns.Range(func(key, value interface{}) bool {
			if value.(string) == name {
				key.(net.Conn).Close()
				h.userconns.Delete(key)
			}
			return true
		})
	}
	// Build new lists excluding deleted users
	remainingNames := make([]string, 0, len(h.userNameList))
	remainingPasswords := make([]string, 0, len(h.userPasswordList))
	for i, user := range h.userNameList {
		if _, found := toDelete[user]; !found {
			remainingNames = append(remainingNames, user)
			if i < len(h.userPasswordList) {
				remainingPasswords = append(remainingPasswords, h.userPasswordList[i])
			}
		}
	}
	h.userNameList = remainingNames
	h.userPasswordList = remainingPasswords
	
	var userUUIDList [][16]byte
	indexs := make([]int, len(h.userNameList))
	for i, UUID := range h.userNameList {
		indexs[i] = h.uuidToUid[UUID]
		userUUID, err := uuid.FromString(UUID)
		if err != nil {
			return E.Cause(err, "invalid uuid for user ", i)
		}
		userUUIDList = append(userUUIDList, userUUID)
	}
	h.server.UpdateUsers(indexs, userUUIDList, h.userPasswordList) // Use password list
	return nil
}
