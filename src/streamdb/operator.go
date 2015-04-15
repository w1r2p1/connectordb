package streamdb

import (
	"streamdb/users"
	"errors"
	"strings"
)

var (
	PERMISSION_ERROR = errors.New("Insufficient Privileges")
	INVALID_PATH_ERROR = errors.New("The given path is invalid")
	InvalidParameterError = errors.New("Invalid Parameter")
)

//Returns the Administrator device (which has all possible permissions)
//Having a nil users.Device means that it is administrator
func (db *Database) GetAdminOperator() *Operator {
	return &Operator{db, nil}
}

//Given an API key, returns the  Device object
func (db *Database) GetOperator(apikey string) (*Operator, error) {
	dev, err := db.ReadDeviceByApiKey(apikey)
	if err != nil {
		return nil, err
	}
	return &Operator{db, dev}, nil
}


// The operatror is a database proxy for a particular device, note that these
// should not be operated indefinitely as the users.Device may change over
// time.
type Operator struct {
	db *Database // the database this operator works on
	dev *users.Device // the device behind this operator
}

// The operating environment for a particular operator request,
// the idea is you can construct one using an operator, then perform a plethora
// of operations using it.
type OperatorRequestEnv struct {
	Operator

	RequestUser   *users.User
	RequestDevice *users.Device
	RequestStream *users.Stream
}

func (o *Operator) GetDevice() (*users.Device) {
	return o.dev
}

func (o *Operator) GetDatabase() (*Database) {
	return o.db
}

// Creates a user with a username, password, and email string
func (o *Operator) CreateUser(username, email, password string) error {
	if ! o.dev.GeneralPermissions().Gte(users.ROOT) {
		return PERMISSION_ERROR
	}

	return o.GetDatabase().CreateUser(username, email, password)
}

func (o *Operator) ReadUser(username string) (*users.User, error) {
	if ! o.dev.GeneralPermissions().Gte(users.ROOT) {
		return nil, PERMISSION_ERROR
	}

	return o.GetDatabase().ReadUserByName(username)
}

func (o *Operator) ReadUserById(id int64) (*users.User, error) {
	if ! o.dev.GeneralPermissions().Gte(users.ROOT) {
		return nil, PERMISSION_ERROR
	}

	return o.GetDatabase().ReadUserById(id)
}


// Returns a User instance if a user exists with the given email address
func (o *Operator) ReadUserByEmail(email string) (*users.User, error) {
	if ! o.dev.GeneralPermissions().Gte(users.ROOT) {
		return nil, PERMISSION_ERROR
	}

	return o.GetDatabase().ReadUserByEmail(email)
}

// Fetches all users from the database
func (o *Operator) ReadAllUsers() ([]users.User, error){
	if ! o.dev.GeneralPermissions().Gte(users.ROOT) {
		return nil, PERMISSION_ERROR
	}

	return o.GetDatabase().ReadAllUsers()
}

// Attempts to update a user as the given device.
func (o *Operator) UpdateUser(user *users.User) error {
	if user == nil {
		return InvalidParameterError
	}

	if ! o.dev.RelationToUser(user).Gte(users.ROOT) {
		return PERMISSION_ERROR
	}

	return o.GetDatabase().UpdateUser(user)
}

// Attempts to delete a user as the given device.
func (o *Operator) DeleteUser(id int64) error {
	if ! o.dev.GeneralPermissions().Gte(users.ROOT) {
		return PERMISSION_ERROR
	}

	return o.GetDatabase().DeleteUser(id)
}

// Attempts to create a phone carrier as the given device
func (o *Operator) CreatePhoneCarrier(name, emailDomain string) error {
	if ! o.dev.GeneralPermissions().Gte(users.ROOT) {
		return PERMISSION_ERROR
	}

	return o.GetDatabase().CreatePhoneCarrier(name, emailDomain)
}

// ReadPhoneCarrierByIdAs attempts to select a phone carrier from the database given its ID
func (o *Operator) ReadPhoneCarrierById(Id int64) (*users.PhoneCarrier, error) {
	if ! o.dev.GeneralPermissions().Gte(users.ENABLED) {
		return nil, PERMISSION_ERROR
	}

	// currently no permissions needed for this
	return o.GetDatabase().ReadPhoneCarrierById(Id)
}

// Attempts to read phone carriers as the given device
func (o *Operator) ReadAllPhoneCarriers() ([]users.PhoneCarrier, error) {
	if ! o.dev.GeneralPermissions().Gte(users.ENABLED) {
		return nil, PERMISSION_ERROR
	}

	return o.GetDatabase().ReadAllPhoneCarriers()
}


// Attempts to update the phone carrier as the given device
func (o *Operator) UpdatePhoneCarrier(carrier *users.PhoneCarrier) error {
	if carrier == nil {
		return InvalidParameterError
	}

	if ! o.dev.GeneralPermissions().Gte(users.ROOT) {
		return PERMISSION_ERROR
	}

	return o.GetDatabase().UpdatePhoneCarrier(carrier)
}

// Attempts to delete the phone carrier as the given device
func (o *Operator) DeletePhoneCarrier(carrierId int64) error {
	if ! o.dev.GeneralPermissions().Gte(users.ROOT) {
		return PERMISSION_ERROR
	}

	return o.GetDatabase().DeletePhoneCarrier(carrierId)
}

func (o *Operator) CreateDevice(Name string, Owner *users.User) error {
	if Owner == nil {
		return InvalidParameterError
	}

	if ! o.dev.RelationToUser(Owner).Gte(users.USER) {
		return PERMISSION_ERROR
	}

	return o.GetDatabase().CreateDevice(Name, Owner.UserId)
}

func (o *Operator) ReadDevicesForUser(u *users.User) ([]users.Device, error) {
	if ! o.dev.RelationToUser(u).Gte(users.FAMILY) {
		return nil, PERMISSION_ERROR
	}

	return o.GetDatabase().ReadDevicesForUserId(u.UserId)
}

func (o *Operator) ReadDeviceByApiKey(Key string) (*users.Device, error) {
	if ! o.dev.GeneralPermissions().Gte(users.ROOT) {
		return nil, PERMISSION_ERROR
	}

	return o.db.ReadDeviceByApiKey(Key)
}

func (o *Operator) UpdateDevice(update *users.Device) error {
	if update == nil {
		return InvalidParameterError
	}

	if ! o.dev.RelationToDevice(update).Gte(users.DEVICE) {
		return PERMISSION_ERROR
	}

	return o.db.UpdateDevice(update)
}

func (o *Operator) DeleteDevice(device *users.Device) error {
	if device == nil {
		return InvalidParameterError
	}

	if ! o.dev.RelationToDevice(device).Gte(users.USER) {
		return PERMISSION_ERROR
	}

	return o.db.DeleteDevice(device.DeviceId)
}

func (o *Operator) CreateStream(Name, Type string, owner *users.Device) (error) {
	if owner == nil {
		return InvalidParameterError
	}

	if ! o.dev.RelationToDevice(owner).Gte(users.USER) {
		return PERMISSION_ERROR
	}

	return o.db.CreateStream(Name, Type, owner.DeviceId)
}

func (o *Operator) ReadStreamsByDevice(operand *users.Device) ([]users.Stream, error) {
	if ! o.dev.RelationToDevice(operand).Gte(users.FAMILY) {
		return nil, PERMISSION_ERROR
	}

	return o.db.ReadStreamsByDevice(operand.DeviceId)
}

func (o *Operator) UpdateStream(d *users.Device, stream *users.Stream) error {

	if ! o.dev.RelationToStream(stream, d).Gte(users.USER) {
		return PERMISSION_ERROR
	}

	return o.db.UpdateStream(stream)
}

func (o *Operator) DeleteStream(d *users.Device, s *users.Stream) error {
	if ! o.dev.RelationToStream(s, d).Gte(users.USER) {
		return PERMISSION_ERROR
	}

	return o.db.DeleteStream(s.StreamId)
}

/**
// Returns a request environment for performing a specific query.
func (o *Operator) GetRequestEnvironment(path string) (ore *OperatorRequestEnv, error) {
	u, d, s, err := ResolvePath(path)

	return &OperatorRequestEnv{o.db, o.dev, u, d, s}, err
}
**/

/**
Converts a path like user/device/stream into the literal user, device and stream

The path may only fill from the left, e.g. "user//" meaning it will only return
the user and nil for the others. Otherwise, the path may fill from the right,
e.g. "/devicename/stream" in which case the user is implicitly the user belonging
to the operator's device.

**/
func (o *Operator) ResolvePath(path string) (user *users.User, device *users.Device, stream *users.Stream, err error) {
	err = nil

	pathsplit := strings.Split(path, "/")
	if len(pathsplit) != 3 {
		return nil, nil, nil, INVALID_PATH_ERROR
	}

	uname := pathsplit[0]
	dname := pathsplit[1]
	sname := pathsplit[2]

	// Parse the user
	if uname == "" {
		user, err = o.ReadUserById(o.GetDevice().UserId)

		if err != nil {
			return user, device, stream, err
		}
	} else {
		user, err = o.ReadUserById(o.GetDevice().UserId)

		if err != nil {
			return user, device, stream, err
		}
	}

	// Parse the device
	if dname == "" {
		device = o.GetDevice()
	} else {
		device, err := o.db.ReadDeviceForUserByName(user.UserId, dname)
		if err != nil {
			return user, device, stream, err
		}
	}

	if sname != "" {
		stream, err = o.db.ReadStreamByDeviceIdAndName(device.DeviceId, sname)
	}

	return user, device, stream, err
}
