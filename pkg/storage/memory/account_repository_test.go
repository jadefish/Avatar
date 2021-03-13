package memory_test

import (
	"testing"
	"time"

	"github.com/jadefish/avatar"
	. "github.com/jadefish/avatar/pkg/storage/memory"
)

func emptyRepo(t *testing.T) avatar.AccountRepository {
	t.Helper()

	db := make(map[avatar.EntityID]*avatar.Account)

	return NewAccountRepo(db)
}

func repoWithAccount(t *testing.T, data avatar.CreateAccountData) (avatar.AccountRepository, avatar.EntityID) {
	t.Helper()

	db := make(map[avatar.EntityID]*avatar.Account)
	id := avatar.EntityID(1)
	now := time.Now()
	account := &avatar.Account{
		ID:           id,
		Name:         data.Name,
		Email:        data.Email,
		PasswordHash: data.PasswordHash,
		CreationIP:   data.CreationIP,
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    nil,
	}
	db[id] = account
	repo := NewAccountRepo(db)

	return repo, id
}

func repoWithDeletedAccount(t *testing.T, data avatar.CreateAccountData) (avatar.AccountRepository, avatar.EntityID) {
	t.Helper()

	db := make(map[avatar.EntityID]*avatar.Account)
	id := avatar.EntityID(1)
	now := time.Now()
	account := &avatar.Account{
		ID:           id,
		Name:         data.Name,
		Email:        data.Email,
		PasswordHash: data.PasswordHash,
		CreationIP:   data.CreationIP,
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    &now,
	}
	db[id] = account
	repo := NewAccountRepo(db)

	return repo, id
}

func TestAccountRepo_Create(t *testing.T) {
	repo := emptyRepo(t)
	data := avatar.CreateAccountData{Name: "meowmix123"}
	err := repo.Create(data)

	if err != nil {
		t.Fatal(err)
	}

	account, err := repo.Get(1)

	if err != nil {
		t.Fatal(err)
	}

	if account == nil || account.Name != data.Name {
		t.Error("Create() did not store a new Account")
	}
}

func TestAccountRepo_Delete(t *testing.T) {
	repo, id := repoWithAccount(t, avatar.CreateAccountData{Name: "meowmix123"})
	err := repo.Delete(id)

	if err != nil {
		t.Fatal(err)
	}

	account, err := repo.Get(id)

	if err != nil {
		t.Fatal(err)
	}

	if account != nil {
		t.Error("Delete() did not delete the Account")
	}
}

func TestAccountRepo_Delete_NoAccount(t *testing.T) {
	repo := emptyRepo(t)
	err := repo.Delete(1)

	if err == nil {
		t.Error("Delete() deleted an Account that does not exist")
	}
}

func TestAccountRepo_Get(t *testing.T) {
	repo, id := repoWithAccount(t, avatar.CreateAccountData{Name: "meowmix123"})
	account, err := repo.Get(id)

	if err != nil {
		t.Fatal(err)
	}

	if account == nil || account.ID != id {
		t.Error("Get() did not retrieve the Account")
	}
}

func TestAccountRepo_Get_DeletedAccount(t *testing.T) {
	repo, id := repoWithDeletedAccount(t, avatar.CreateAccountData{Name: "meowmix123"})
	account, err := repo.Get(id)

	if err != nil {
		t.Fatal(err)
	}

	if account != nil {
		t.Error("Get() retrieved a deleted Account")
	}
}

func TestAccountRepo_Update(t *testing.T) {
	repo, id := repoWithAccount(t, avatar.CreateAccountData{Name: "meowmix123"})
	newName := "updated"
	err := repo.Update(id, avatar.UpdateAccountData{Name: newName})

	if err != nil {
		t.Fatal(err)
	}

	account, err := repo.Get(id)

	if err != nil {
		t.Fatal(err)
	}

	if account.Name != newName {
		t.Error("Update() did not store updated Account data")
	}
}

func TestAccountRepo_Update_NoAccount(t *testing.T) {
	repo := emptyRepo(t)
	err := repo.Update(1, avatar.UpdateAccountData{Name: "meowmix123"})

	if err == nil {
		t.Error("Update() updated an Account that does not exist")
	}
}

func TestAccountRepo_FindByEmail(t *testing.T) {
	email := "email@example.com"
	repo, _ := repoWithAccount(t, avatar.CreateAccountData{Email: email})

	account, err := repo.FindByEmail(email)

	if err != nil {
		t.Fatal(err)
	}

	if account == nil || account.Email != email {
		t.Error("FindByEmail() did not retrieve the Account")
	}
}

func TestAccountRepo_FindByEmail_DeletedAccount(t *testing.T) {
	email := "email@example.com"
	repo, _ := repoWithDeletedAccount(t, avatar.CreateAccountData{Email: email})
	account, err := repo.FindByEmail(email)

	if err != nil {
		t.Fatal(err)
	}

	if account != nil {
		t.Error("FindByEmail() retrieved a deleted Account")
	}
}

func TestAccountRepo_FindByName(t *testing.T) {
	name := "my_account"
	repo, _ := repoWithAccount(t, avatar.CreateAccountData{Name: name})
	account, err := repo.FindByName(name)

	if err != nil {
		t.Fatal(err)
	}

	if account == nil || account.Name != name {
		t.Error("FindByName() did not retrieve the Account")
	}
}
