package cookie_clicker

import (
	"time"
)

type GameStateData struct {
	NCookies      float64              `json:"n_cookies"`
	NBuildings    map[BuildingType]int `json:"n_buildings"`
	UpgradeStatus map[UpgradeID]bool   `json:"upgrade_status"`
}

type GameStateStruct struct {
	/* Imported Fields */
	nCookies      float64
	nBuildings    map[BuildingType]int
	upgradeStatus map[UpgradeID]bool
	/* Calculated Cache Fields */
	cookiesPerClick float64
	cps             float64
	/* Immutable Fields */
	buildingCPSRef     map[BuildingType]float64
	cookiesPerClickRef float64
	buildingCost       map[BuildingType]BuildingCostFunction
	upgrades           map[UpgradeID]UpgradeInterface
}

func NewGameState() *GameStateStruct {
	g := GameStateStruct{
		nBuildings:     make(map[BuildingType]int),
		buildingCPSRef: make(map[BuildingType]float64),
		buildingCost:   make(map[BuildingType]BuildingCostFunction),
		upgradeStatus:  make(map[UpgradeID]bool),
		upgrades:       make(map[UpgradeID]UpgradeInterface),
	}

	var i BuildingType
	for i = 0; i < BUILDING_TYPE_ENUM_EOF; i++ {
		g.nBuildings[i] = 0
		g.buildingCPSRef[i] = 0
		g.buildingCost[i] = func(current int) float64 { return 0 }
	}

	var j UpgradeID
	for j = 0; j < UPGRADE_ID_ENUM_EOF; j++ {
		g.upgradeStatus[j] = false
	}

	return &g
}

/* Public API */

func (self *GameStateStruct) Load(d GameStateData) {
	self.loadData(d)
	self.loadBuildingCost(BUILDING_COST_LOOKUP)
	self.loadUpgrades(UPGRADES_LOOKUP)
	self.loadBuildingCPSRef(BUILDING_CPS_LOOKUP)
	self.loadCookiesPerClickRef(COOKIES_PER_CLICK_LOOKUP)
}

func (self *GameStateStruct) GetNBuildings() map[BuildingType]int {
	return self.nBuildings
}

func (self *GameStateStruct) GetBuildingCost() map[BuildingType]BuildingCostFunction {
	return self.buildingCost
}

func (self *GameStateStruct) GetUpgrades() map[UpgradeID]UpgradeInterface {
	return self.upgrades
}

func (self *GameStateStruct) GetUpgradeStatus() map[UpgradeID]bool {
	return self.upgradeStatus
}

func (self *GameStateStruct) GetCookies() float64 {
	return self.nCookies
}

func (self *GameStateStruct) BuyUpgrade(id UpgradeID) bool { // TODO(cripplet): Enforce upgrade cost check.
	upgrade, present := self.upgrades[id]
	to_buy := present && !self.upgradeStatus[id]
	bought := to_buy && upgrade.GetIsUnlocked(self) && self.subtractCookies(upgrade.GetCost(self))
	if bought {
		self.upgradeStatus[id] = true
		self.setCPS(self.calculateCPS())
		self.setCookiesPerClick(self.calculateCookiesPerClick())

	}
	return bought
}

func (self *GameStateStruct) BuyBuilding(buildingType BuildingType) bool {
	cost := self.buildingCost[buildingType](self.nBuildings[buildingType])
	bought := self.subtractCookies(cost)
	if bought {
		self.nBuildings[buildingType] += 1
		self.setCPS(self.calculateCPS())
		self.setCookiesPerClick(self.calculateCookiesPerClick())
	}
	return bought
}

func (self *GameStateStruct) GetCPS(start time.Time, end time.Time) float64 { // TODO(cripplet): Calculate timed buffs here.
	return self.cps * float64(end.Sub(start)) / float64(time.Second)
}

func (self *GameStateStruct) GetCookiesPerClick() float64 { // TODO(cripplet): Calculate timed buffs here.
	return self.cookiesPerClick
}

func (self *GameStateStruct) MineCookies(start time.Time, end time.Time) {
	self.addCookies(self.GetCPS(start, end))
}

func (self *GameStateStruct) Click() {
	self.addCookies(self.GetCookiesPerClick())
}

/* End public API */

func (self *GameStateStruct) setCPS(cps float64) {
	self.cps = cps
}

func (self *GameStateStruct) addCookies(n float64) {
	self.nCookies += n
}

func (self *GameStateStruct) subtractCookies(n float64) bool {
	if self.nCookies >= n {
		self.nCookies -= n
		return true
	}
	return false
}

func (self *GameStateStruct) setCookiesPerClick(c float64) {
	self.cookiesPerClick = c
}

func (self *GameStateStruct) loadData(d GameStateData) {
	self.nCookies = d.NCookies
	self.nBuildings = d.NBuildings
	self.upgradeStatus = d.UpgradeStatus
}

func (self *GameStateStruct) loadBuildingCost(c map[BuildingType]BuildingCostFunction) {
	for buildingType := range self.buildingCost {
		self.buildingCost[buildingType] = func(current int) float64 { return 0 }
	}

	for buildingType, buildingCostFunction := range c {
		self.buildingCost[buildingType] = buildingCostFunction
	}
}

func (self *GameStateStruct) loadUpgrades(u map[UpgradeID]UpgradeInterface) {
	for upgradeID := range self.upgrades {
		delete(self.upgrades, upgradeID)
	}

	for upgradeID, upgradeInterface := range u {
		self.upgrades[upgradeID] = upgradeInterface
	}
}

func (self *GameStateStruct) loadCookiesPerClickRef(c float64) {
	self.cookiesPerClickRef = c
}

func (self *GameStateStruct) loadBuildingCPSRef(b map[BuildingType]float64) {
	for buildingType := range self.buildingCPSRef {
		self.buildingCPSRef[buildingType] = 0
	}

	for buildingType, buildingCPSRef := range b {
		self.buildingCPSRef[buildingType] = buildingCPSRef
	}
}

func (self *GameStateStruct) calculateCookiesPerClick() float64 {
	cookiesPerClickCopy := self.cookiesPerClickRef

	for upgradeID, bought := range self.upgradeStatus {
		if bought {
			cookiesPerClickCopy *= self.GetUpgrades()[upgradeID].GetClickMultiplier(self)
		}
	}

	return cookiesPerClickCopy
}

func (self *GameStateStruct) calculateCPS() float64 {
	buildingCPSRefCopy := make(map[BuildingType]float64)
	for buildingType, buildingTypeCPS := range self.buildingCPSRef {
		buildingCPSRefCopy[buildingType] = buildingTypeCPS
	}

	boughtUpgrades := make([]UpgradeInterface, 0)
	for upgradeID, upgrade := range self.upgrades {
		if self.upgradeStatus[upgradeID] {
			boughtUpgrades = append(boughtUpgrades, upgrade)
		}
	}

	for _, upgrade := range boughtUpgrades {
		if upgrade.GetBuildingType() < BUILDING_TYPE_ENUM_EOF {
			buildingCPSRefCopy[upgrade.GetBuildingType()] *= upgrade.GetBuildingMultiplier(self)
		}
	}

	var totalCPS float64
	for buildingType, cps := range buildingCPSRefCopy {
		totalCPS += float64(self.nBuildings[buildingType]) * cps
	}

	return totalCPS
}
