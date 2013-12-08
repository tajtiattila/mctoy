package net

// StatePlay
////////////////////////////////////////////////////////////////////////////////

// 0x01 = Chat Message
type ClientChatMessage struct {
	Message string
}

func (ClientChatMessage) Id() (uint, uint) { return PktInvalid, 0x01 }

// 0x02 = Use Entity
type UseEntity struct {
	Target int32
	Mouse  int8 // 0 = Left-click, 1 = Right-click
}

func (UseEntity) Id() (uint, uint) { return PktInvalid, 0x02 }

// 0x03 = Player
type Player struct {
	OnGround bool // True if the client is on the ground, False otherwise
}

func (Player) Id() (uint, uint) { return PktInvalid, 0x03 }

// 0x04 = Player Position
type PlayerPosition struct {
	X        float64 // Absolute position
	Y        float64 // Absolute position
	Stance   float64 // Used to modify the players bounding box when going up stairs, crouching, etc…
	Z        float64 // Absolute position
	OnGround bool    // True if the client is on the ground, False otherwise
}

func (PlayerPosition) Id() (uint, uint) { return PktInvalid, 0x04 }

// 0x05 = Player Look
type PlayerLook struct {
	Yaw      float32 // Absolute rotation on the X Axis, in degrees
	Pitch    float32 // Absolute rotation on the Y Axis, in degrees
	OnGround bool    // True if the client is on the ground, False otherwise
}

func (PlayerLook) Id() (uint, uint) { return PktInvalid, 0x05 }

// 0x06 = Player Position And Look
type ClientPlayerPositionAndLook struct {
	X        float64 // Absolute position
	Y        float64 // Absolute position
	Stance   float64 // Used to modify the players bounding box when going up stairs, crouching, etc…
	Z        float64 // Absolute position
	Yaw      float32 // Absolute rotation on the X Axis, in degrees
	Pitch    float32 // Absolute rotation on the Y Axis, in degrees
	OnGround bool    // True if the client is on the ground, False otherwise
}

func (ClientPlayerPositionAndLook) Id() (uint, uint) { return PktInvalid, 0x06 }

// 0x07 = Player Digging
type PlayerDigging struct {
	Status int8  // The action the player is taking against the block (see below)
	X      int32 // Block position
	Y      uint8 // Block position
	Z      int32 // Block position
	Face   int8  // The face being hit (see below)
}

func (PlayerDigging) Id() (uint, uint) { return PktInvalid, 0x07 }

// 0x08 = Player Block Placement
type PlayerBlockPlacement struct {
	X               int32 // Block position
	Y               uint8 // Block position
	Z               int32 // Block position
	Direction       int8  // The offset to use for block/item placement (see below)
	HeldItem        Slot
	CursorPositionX int8 // The position of the crosshair on the block
	CursorPositionY int8
	CursorPositionZ int8
}

func (PlayerBlockPlacement) Id() (uint, uint) { return PktInvalid, 0x08 }

// 0x09 = Held Item Change
type ClientHeldItemChange struct {
	Slot int16 // The slot which the player has selected (0-8)
}

func (ClientHeldItemChange) Id() (uint, uint) { return PktInvalid, 0x09 }

// 0x0A = Animation
type ClientAnimation struct {
	EntityID  int32 // Player ID
	Animation int8  // Animation ID
}

func (ClientAnimation) Id() (uint, uint) { return PktInvalid, 0x0A }

// 0x0B = Entity Action
type EntityAction struct {
	EntityID  int32 // Player ID
	ActionID  int8  // The ID of the action, see below.
	JumpBoost int32 // Horse jump boost. Ranged from 0 -> 100.
}

func (EntityAction) Id() (uint, uint) { return PktInvalid, 0x0B }

// 0x0C = Steer Vehicle
type SteerVehicle struct {
	Sideways float32 // Positive to the left of the player
	Forward  float32 // Positive forward
	Jump     bool
	Unmount  bool // True when leaving the vehicle
}

func (SteerVehicle) Id() (uint, uint) { return PktInvalid, 0x0C }

// 0x0E = Click Window
type ClickWindow struct {
	WindowID     int8  // The id of the window which was clicked. 0 for player inventory.
	Slot         int16 // The clicked slot. See below.
	Button       int8  // The button used in the click. See below.
	ActionNumber int16 // A unique number for the action, used for transaction handling (See the Transaction packet).
	Mode         int8  // Inventory operation mode. See below.
	ClickedItem  Slot
}

func (ClickWindow) Id() (uint, uint) { return PktInvalid, 0x0E }

// 0x10 = Creative Inventory Action
type CreativeInventoryAction struct {
	Slot        int16 // Inventory slot
	ClickedItem Slot
}

func (CreativeInventoryAction) Id() (uint, uint) { return PktInvalid, 0x10 }

// 0x11 = Enchant Item
type EnchantItem struct {
	WindowID    int8 // The ID sent by [[#0x64|Open Window]]
	Enchantment int8 // The position of the enchantment on the enchantment table window, starting with 0 as the topmost one.
}

func (EnchantItem) Id() (uint, uint) { return PktInvalid, 0x11 }

// 0x14 = Tab-Complete
type TabCompleteRequest struct {
	Text string
}

func (TabCompleteRequest) Id() (uint, uint) { return PktInvalid, 0x14 }

// 0x15 = Client Settings
type ClientSettings struct {
	Locale       string // en_GB
	ViewDistance int8   // 0-3 for 'far', 'normal', 'short', 'tiny'.
	ChatFlags    int8   // Chat settings. See notes below.
	Unused       bool   // Only observed as true
	Difficulty   int8   // Client-side difficulty from options.txt
	ShowCape     bool   // Client-side "show cape" option
}

func (ClientSettings) Id() (uint, uint) { return PktInvalid, 0x15 }

// 0x16 = Client Status
type ClientStatus struct {
	ActionID int8 // See below
}

func (ClientStatus) Id() (uint, uint) { return PktInvalid, 0x16 }

////////////////////////////////////////////////////////////////////////////////
