package dominos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/cptaffe/pizza/pizza"
)

type NearStore struct {
	StoreID                 string
	IsDeliveryStore         bool
	MinDistance             float32
	MaxDistance             float32
	Phone                   string
	AddressDescription      string
	HolidaysDescription     string
	HoursDescription        string
	ServiceHoursDescription struct {
		Carryout string
		Delivery string
	}
	IsOnlineCapable   bool
	IsOnlineNow       bool
	IsNEONow          bool
	IsSpanish         bool
	SubstitutionStore string
	// TODO: figure out what these are:
	LocationInfo         string            // was null
	LanguageLocationInfo map[string]string // was null
	AllowDeliveryOrders  bool
	AllowCarryoutOrders  bool
	IsOpen               bool
	ServiceIsOpen        struct {
		Carryout bool
		Delivery bool
	}
}

type NearResp struct {
	Status      int
	Granularity string
	Address     struct {
		Street       string
		StreetNumber string
		StreetName   string
		UnitType     string
		UnitNumber   string
		City         string
		Region       string
		PostalCode   string
	}
	Stores []NearStore
}

// Returns stores sorted by distance
func nearStores(addr *pizza.Addr) ([]NearStore, error) {
	u, err := url.Parse("https://order.dominos.com/power/store-locator")
	if err != nil {
		return []NearStore{}, err
	}
	v := url.Values{}
	v.Add("type", "Locations")
	v.Add("c", fmt.Sprintf("%s, %s %d", addr.City, addr.State, addr.Zip))
	v.Add("s", addr.Street)
	u.RawQuery = v.Encode()
	fmt.Println(u.String())
	r, err := http.Get(u.String())
	if err != nil {
		return []NearStore{}, err
	}
	rsp := NearResp{}
	if err = json.NewDecoder(r.Body).Decode(&rsp); err != nil {
		return rsp.Stores, err
	}
	return rsp.Stores, nil
}

type StoreHours struct {
	OpenTime  string
	CloseTime string
}

type StoreTimes struct {
	Sun []StoreHours
	Mon []StoreHours
	Tue []StoreHours
	Wed []StoreHours
	Thu []StoreHours
	Fri []StoreHours
	Sat []StoreHours
}

type Store struct {
	StoreID                         string
	BusinessDate                    string
	PulseVersion                    string
	PulseVersionName                string
	PreferredLanguage               string
	PreferredCurrency               string
	Phone                           string
	StreetName                      string
	City                            string
	Region                          string
	PostalCode                      string
	AddressDescription              string
	TimeZoneCode                    string
	TimeZoneMinutes                 int
	IsAffectedByDaylightSavingsTime bool
	Holidays                        struct{}
	HolidayDescription              string
	Hours                           StoreTimes
	HoursDescription                string
	ServiceHours                    struct {
		Carryout StoreTimes
		Delivery StoreTimes
	}
	ServiceHoursDescription struct {
		Carryout string
		Delivery string
	}
	CustomerCloseWarningMinutes  int
	AcceptablePaymentTypes       []string // e.g. Cash, GiftCard, CreditCard
	AcceptableCreditCards        []string // e.g. American Express, Discover Card, Mastercard, Optima, Visa
	IsOnlineCapable              bool
	LocationInfo                 string            // was null
	LanguageLocationInfo         map[string]string // was null
	SubstitutionStore            string
	MinimumDeliveryOrderAmount   float32
	EstimatedCarroutWaitMinutes  string
	CashLimit                    int
	IsForceOffline               bool
	IsOnlineNow                  bool
	IsForceClose                 bool
	IsOpen                       bool
	OnlineStatusCode             string // e.g. Ok
	StoreAsOfTime                string
	AsOfTime                     string
	IsNEONow                     bool
	IsSpanish                    bool
	AllowCarroutOrders           bool
	AllowDeliveryOrders          bool
	Status                       int
	AcceptableWalletTypes        []string          // e.g. Google
	SocialReviewLinks            map[string]string // e.g plus:<url>
	IsAVSEnabled                 bool
	Pop                          bool
	LanguageTranslations         string // was null
	StoreLocation                string // e.g. N/A
	DriverTrackingSupported      bool
	IsCookingInstructionsEnabled bool
	IsSaltWarningEnabled         bool
	EstimatedWaitMinutes         string // e.g. 20-30
	Metadata                     string
}

// Returns Store from StoreID
func storeFromID(id int) (pizza.Store, error) {
	u := fmt.Sprintf("https://order.dominos.com/power/store/%d/profile", id)
	fmt.Println(u)
	r, err := http.Get(u)
	if err != nil {
		return &Store{}, err
	}
	store := new(Store)
	if err = json.NewDecoder(r.Body).Decode(store); err != nil {
		return store, err
	}
	return store, err
}

func (s *Store) Addr() (pizza.Addr, error) {
	return pizza.Addr{
		Street: s.StreetName,
		City:   s.City,
		State:  s.Region,
		Zip:    s.PostalCode,
	}, nil
}

func Stores(addr *pizza.Addr) ([]pizza.Store, error) {
	stores := make(chan pizza.Store)
	ns, err := nearStores(addr)
	if err != nil {
		return []pizza.Store{}, err
	}
	var wg sync.WaitGroup
	for _, s := range ns {
		id, err := strconv.Atoi(s.StoreID)
		if err != nil {
			return []pizza.Store{}, err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			st, err := storeFromID(id)
			if err != nil {
				return
			}
			stores <- st
		}()
	}
	go func() {
		wg.Wait()
		close(stores)
	}()
	sts := []pizza.Store{}
	for s := range stores {
		sts = append(sts, s)
	}
	return sts, nil
}
