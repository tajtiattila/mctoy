package net

// StatePlay
////////////////////////////////////////////////////////////////////////////////

type KeepAlive struct {
	// 0x00 ->Server ->Client
	KeepAliveID int32
}

func (KeepAlive) Id() (uint, uint) { return 0x00, 0x00 }

// 0x01 = Join Game
type JoinGame struct {
	EntityID   int32  // The player's Entity ID
	Gamemode   uint8  // 0: survival, 1: creative, 2: adventure. Bit 3 (0x8) is the hardcore flag
	Dimension  int8   // -1: nether, 0: overworld, 1: end
	Difficulty uint8  // 0 thru 3 for Peaceful, Easy, Normal, Hard
	MaxPlayers uint8  // Used by the client to draw the player list
	LevelType  string // default, flat, largeBiomes, amplified, default_1_1
}

func (JoinGame) Id() (uint, uint) { return 0x01, PktInvalid }

// 0x02 = Chat Message
type ServerChatMessage struct {
	JSONData string // https://gist.github.com/thinkofdeath/e882ce057ed83bac0a1c , Limited to 32767 bytes
}

func (ServerChatMessage) Id() (uint, uint) { return 0x02, PktInvalid }

// 0x03 = Time Update
type TimeUpdate struct {
	AgeOfTheWorld int64 // In ticks; not changed by server commands
	TimeOfDay     int64 // The world (or region) time, in ticks. If negative the sun will stop moving at the Math.abs of the time
}

func (TimeUpdate) Id() (uint, uint) { return 0x03, PktInvalid }

// 0x04 = Entity Equipment
type EntityEquipment struct {
	EntityID int32 // Entity's ID
	Slot     int16 // Equipment slot: 0=held, 1-4=armor slot (1 - boots, 2 - leggings, 3 - chestplate, 4 - helmet)
	Item     Slot  // Item in slot format
}

func (EntityEquipment) Id() (uint, uint) { return 0x04, PktInvalid }

// 0x05 = Spawn Position
type SpawnPosition struct {
	X int32 // Spawn X in block coordinates
	Y int32 // Spawn Y in block coordinates
	Z int32 // in block coordinates
}

func (SpawnPosition) Id() (uint, uint) { return 0x05, PktInvalid }

// 0x06 = Update Health
type UpdateHealth struct {
	Health         float32 // 0 or less = dead, 20 = full HP
	Food           int16   // 0 - 20
	FoodSaturation float32 // Seems to vary from 0.0 to 5.0 in integer increments
}

func (UpdateHealth) Id() (uint, uint) { return 0x06, PktInvalid }

// 0x07 = Respawn
type Respawn struct {
	Dimension  int32  // -1: The Nether, 0: The Overworld, 1: The End
	Difficulty uint8  // 0 thru 3 for Peaceful, Easy, Normal, Hard.
	Gamemode   uint8  // 0: survival, 1: creative, 2: adventure. The hardcore flag is not included
	LevelType  string // Same as [[Protocol#Join_Game|Join Game]]
}

func (Respawn) Id() (uint, uint) { return 0x07, PktInvalid }

// 0x08 = Player Position And Look (Clientbound)
type ServerPlayerPositionAndLook struct {
	X        float64 // Absolute position
	Y        float64 // Absolute position
	Z        float64 // Absolute position
	Yaw      float32 // Absolute rotation on the X Axis, in degrees
	Pitch    float32 // Absolute rotation on the Y Axis, in degrees
	OnGround bool    // True if the client is on the ground, False otherwise
}

func (ServerPlayerPositionAndLook) Id() (uint, uint) { return 0x08, PktInvalid }

// 0x09 = Held Item Change
type ServerHeldItemChange struct {
	Slot int8 // The slot which the player has selected (0-8)
}

func (ServerHeldItemChange) Id() (uint, uint) { return 0x09, PktInvalid }

// 0x0A = Use Bed
type UseBed struct {
	EntityID int32 // Player ID
	X        int32 // Bed headboard X as block coordinate
	Y        uint8 // Bed headboard Y as block coordinate
	Z        int32 // Bed headboard Z as block coordinate
}

func (UseBed) Id() (uint, uint) { return 0x0A, PktInvalid }

// 0x0B = Animation
type ServerAnimation struct {
	EntityID  uint  // Player ID
	Animation uint8 // Animation ID
}

func (ServerAnimation) Id() (uint, uint) { return 0x0B, PktInvalid }

// 0x0C = Spawn Player
type SpawnPlayer struct {
	EntityID    uint       // Player's Entity ID
	PlayerUUID  string     // Player's UUID
	PlayerName  string     // Player's Name
	X           int32      // Player X as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Y           int32      // Player X as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Z           int32      // Player X as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Yaw         int8       // Player rotation as a packed byte
	Pitch       int8       // Player rotation as a packet byte
	CurrentItem int16      // The item the player is currently holding. Note that this should be 0 for "no item", unlike -1 used in other packets. A negative value crashes clients.
	Metadata    EntityData // The client will crash if no metadata is sent
}

func (SpawnPlayer) Id() (uint, uint) { return 0x0C, PktInvalid }

// 0x0D = Collect Item
type CollectItem struct {
	CollectedEntityID int32
	CollectorEntityID int32
}

func (CollectItem) Id() (uint, uint) { return 0x0D, PktInvalid }

// 0x0E = Spawn Object
type SpawnObject struct {
	EntityID uint  // Entity ID of the object
	Type     int8  // The type of object (See [[Entities#Objects|Objects]])
	X        int32 // X position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Y        int32 // Y position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Z        int32 // Z position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Pitch    int8  // The pitch in steps of 2p/256
	Yaw      int8  // The yaw in steps of 2p/256
	Data     ObjectData
}

func (SpawnObject) Id() (uint, uint) { return 0x0E, PktInvalid }

// 0x0F = Spawn Mob
type SpawnMob struct {
	EntityID  uint  // Entity's ID
	Type      uint8 // The type of mob. See [[Entities#Mobs|Mobs]]
	X         int32 // X position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Y         int32 // Y position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Z         int32 // Z position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Pitch     int8  // The pitch in steps of 2p/256
	HeadPitch int8  // The pitch in steps of 2p/256
	Yaw       int8  // The yaw in steps of 2p/256
	VelocityX int16
	VelocityY int16
	VelocityZ int16
	Metadata  EntityData
}

func (SpawnMob) Id() (uint, uint) { return 0x0F, PktInvalid }

// 0x10 = Spawn Painting
type SpawnPainting struct {
	EntityID  uint   // Entity's ID
	Title     string // Name of the painting. Max length 13
	X         int32  // Center X coordinate
	Y         int32  // Center Y coordinate
	Z         int32  // Center Z coordinate
	Direction int32  // Direction the painting faces (0 -z, 1 -x, 2 +z, 3 +x)
}

func (SpawnPainting) Id() (uint, uint) { return 0x10, PktInvalid }

// 0x11 = Spawn Experience Orb
type SpawnExperienceOrb struct {
	EntityID uint  // Entity's ID
	X        int32 // X position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Y        int32 // Y position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Z        int32 // Z position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Count    int16 // The amount of experience this orb will reward once collected
}

func (SpawnExperienceOrb) Id() (uint, uint) { return 0x11, PktInvalid }

// 0x12 = Entity Velocity
type EntityVelocity struct {
	EntityID  int32 // Entity's ID
	VelocityX int16 // Velocity on the X axis
	VelocityY int16 // Velocity on the Y axis
	VelocityZ int16 // Velocity on the Z axis
}

func (EntityVelocity) Id() (uint, uint) { return 0x12, PktInvalid }

// 0x13 = Destroy Entities
type DestroyEntities struct {
	EntityIDs []uint32 `mc:"len=int8"` // The list of entities of destroy
}

func (DestroyEntities) Id() (uint, uint) { return 0x13, PktInvalid }

// 0x14 = Entity
type Entity struct {
	EntityID int32 // Entity's ID
}

func (Entity) Id() (uint, uint) { return 0x14, PktInvalid }

// 0x15 = Entity Relative Move
type EntityRelativeMove struct {
	EntityID int32 // Entity's ID
	DX       int8  // Change in X position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	DY       int8  // Change in Y position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	DZ       int8  // Change in Z position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
}

func (EntityRelativeMove) Id() (uint, uint) { return 0x15, PktInvalid }

// 0x16 = Entity Look
type EntityLook struct {
	EntityID int32 // Entity's ID
	Yaw      int8  // The X Axis rotation as a fraction of 360
	Pitch    int8  // The Y Axis rotation as a fraction of 360
}

func (EntityLook) Id() (uint, uint) { return 0x16, PktInvalid }

// 0x17 = Entity Look and Relative Move
type EntityLookAndRelativeMove struct {
	EntityID int32 // Entity's ID
	DX       int8  // Change in X position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	DY       int8  // Change in Y position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	DZ       int8  // Change in Z position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Yaw      int8  // The X Axis rotation as a fraction of 360
	Pitch    int8  // The Y Axis rotation as a fraction of 360
}

func (EntityLookAndRelativeMove) Id() (uint, uint) { return 0x17, PktInvalid }

// 0x18 = Entity Teleport
type EntityTeleport struct {
	EntityID int32 // Entity's ID
	X        int32 // X position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Y        int32 // Y position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Z        int32 // Z position as a [[Data_Types#Fixed-point_numbers|Fixed-Point number]]
	Yaw      int8  // The X Axis rotation as a fraction of 360
	Pitch    int8  // The Y Axis rotation as a fraction of 360
}

func (EntityTeleport) Id() (uint, uint) { return 0x18, PktInvalid }

// 0x19 = Entity Head Look
type EntityHeadLook struct {
	EntityID int32 // Entity's ID
	HeadYaw  int8  // Head yaw in steps of 2p/256
}

func (EntityHeadLook) Id() (uint, uint) { return 0x19, PktInvalid }

// 0x1A = Entity Status
type EntityStatus struct {
	EntityID     int32 // Entity's ID
	EntityStatus int8  // See below
}

func (EntityStatus) Id() (uint, uint) { return 0x1A, PktInvalid }

// 0x1B = Attach Entity
type AttachEntity struct {
	EntityID  int32 // Entity's ID
	VehicleID int32 // Vechicle's Entity ID
	Leash     bool  // If true leashes the entity to the vehicle
}

func (AttachEntity) Id() (uint, uint) { return 0x1B, PktInvalid }

// 0x1C = Entity Metadata
type EntityMetadata struct {
	EntityID int32 // Entity's ID
	Metadata EntityData
}

func (EntityMetadata) Id() (uint, uint) { return 0x1C, PktInvalid }

// 0x1D = Entity Effect
type EntityEffect struct {
	EntityID  int32 // Entity's ID
	EffectID  int8  // See [[http://www.minecraftwiki.net/wiki/Potion_effect#Parameters]]
	Amplifier int8
	Duration  int16
}

func (EntityEffect) Id() (uint, uint) { return 0x1D, PktInvalid }

// 0x1E = Remove Entity Effect
type RemoveEntityEffect struct {
	EntityID int32 // Entity's ID
	EffectID int8
}

func (RemoveEntityEffect) Id() (uint, uint) { return 0x1E, PktInvalid }

// 0x1F = Set Experience
type SetExperience struct {
	ExperienceBar   float32 // Between 0 and 1
	Level           int16
	TotalExperience int16
}

func (SetExperience) Id() (uint, uint) { return 0x1F, PktInvalid }

// 0x20 = Entity Properties
type EntityProperties struct {
	EntityID   int32          // Entity's ID
	Properties []PropertyData `mc:"len=int32"`
}

func (EntityProperties) Id() (uint, uint) { return 0x20, PktInvalid }

// 0x21 = Chunk Data
type ChunkData struct {
	ChunkX             int32  // Chunk X coordinate
	ChunkZ             int32  // Chunk Z coordinate
	GroundUpContinuous bool   // This is True if the packet represents all sections in this vertical column, where the primary bit map specifies exactly which sections are included, and which are air
	PrimaryBitMap      int16  // Bitmask with 1 for every 16x16x16 section which data follows in the compressed data.
	AddBitMap          int16  // Same as above, but this is used exclusively for the 'add' portion of the payload
	CompressedData     []byte `mc:"len=int32"` // The chunk data is compressed using Zlib Deflate
}

func (ChunkData) Id() (uint, uint) { return 0x21, PktInvalid }

// 0x22 = Multi Block Change
type MultiBlockChange struct {
	ChunkX      int32    // Chunk X coordinate
	ChunkZ      int32    // Chunk Z Coordinate
	RecordCount int16    // The number of blocks affected
	Records     []Record `mc:"len=int32div4"` // The total size of the data is in bytes. Should always be 4*record count
}

func (MultiBlockChange) Id() (uint, uint) { return 0x22, PktInvalid }

// 0x23 = Block Change
type BlockChange struct {
	X         int32 // Block X Coordinate
	Y         uint8 // Block Y Coordinate
	Z         int32 // Block Z Coordinate
	BlockType uint  // The new block type for the block
	BlockData uint8 // The new data for the block
}

func (BlockChange) Id() (uint, uint) { return 0x23, PktInvalid }

// 0x24 = Block Action
type BlockAction struct {
	X         int32 // Block X Coordinate
	Y         int16 // Block Y Coordinate
	Z         int32 // Block Z Coordinate
	Byte1     uint8 // Varies depending on block - see [[Block_Actions]]
	Byte2     uint8 // Varies depending on block - see [[Block_Actions]]
	BlockType uint  // The block type for the block
}

func (BlockAction) Id() (uint, uint) { return 0x24, PktInvalid }

// 0x25 = Block Break Animation
type BlockBreakAnimation struct {
	EntityID     uint  // Entity's ID
	X            int32 // Block Position
	Y            int32 // Block Position
	Z            int32 // Block Position
	DestroyStage int8  // 0 - 9
}

func (BlockBreakAnimation) Id() (uint, uint) { return 0x25, PktInvalid }

// 0x26 = Map Chunk Bulk
type MapChunkBulk struct {
	ChunkColumnCount int16 // The number of chunk in this packet
	//DataLength            int32      // The size of the data field
	SkyLightSent bool   // Whether or not the chunk data contains a light nibble array. This is true in the main world, false in the end + nether
	Data         []byte // Compressed chunk data
	Meta         MapChunkBulkMeta
}

func (MapChunkBulk) Id() (uint, uint) { return 0x26, PktInvalid }
func (p *MapChunkBulk) MarshalPacket(k *PacketEncoder) {
	k.PutInt16(p.ChunkColumnCount)
	k.PutUint32(uint32(len(p.Data)))
	k.PutBool(p.SkyLightSent)
	d := k.Get(len(p.Data))
	if d != nil {
		copy(d, p.Data)
	}
	k.Encode(&p.Meta)
}
func (p *MapChunkBulk) UnmarshalPacket(k *PacketDecoder) {
	p.ChunkColumnCount = k.Int16()
	dlen := int(k.Int32())
	p.SkyLightSent = k.Bool()
	p.Data = k.Get(dlen)
	k.Decode(&p.Meta)
}

// 0x27 = Explosion
type Explosion struct {
	X             float32
	Y             float32
	Z             float32
	Radius        float32 // Currently unused in the client
	Records       []XYZ8  `mc:"len=int32"` // Each record is 3 signed bytes long, each bytes are the XYZ (respectively) offsets of affected blocks.
	PlayerMotionX float32 // X velocity of the player being pushed by the explosion
	PlayerMotionY float32 // Y velocity of the player being pushed by the explosion
	PlayerMotionZ float32 // Z velocity of the player being pushed by the explosion
}

func (Explosion) Id() (uint, uint) { return 0x27, PktInvalid }

// 0x28 = Effect
type Effect struct {
	EffectID              int32 // The ID of the effect, see below.
	X                     int32 // The X location of the effect
	Y                     int8  // The Y location of the effect
	Z                     int32 // The Z location of the effect
	Data                  int32 // Extra data for certain effects, see below.
	DisableRelativeVolume bool  // See above
}

func (Effect) Id() (uint, uint) { return 0x28, PktInvalid }

// 0x29 = Sound Effect
type SoundEffect struct {
	SoundName       string
	EffectPositionX int32   // Effect X multiplied by 8
	EffectPositionY int32   // Effect Y multiplied by 8
	EffectPositionZ int32   // Effect Z multiplied by 8
	Volume          float32 // 1 is 100%, can be more
	Pitch           uint8   // 63 is 100%, can be more
}

func (SoundEffect) Id() (uint, uint) { return 0x29, PktInvalid }

// 0x2A = Particle
type Particle struct {
	ParticleName      string  // The name of the particle to create. A list can be found [https://gist.github.com/thinkofdeath/5110835 here]
	X                 float32 // X position of the particle
	Y                 float32 // Y position of the particle
	Z                 float32 // Z position of the particle
	OffsetX           float32 // This is added to the X position after being multiplied by random.nextGaussian()
	OffsetY           float32 // This is added to the Y position after being multiplied by random.nextGaussian()
	OffsetZ           float32 // This is added to the Z position after being multiplied by random.nextGaussian()
	ParticleData      float32 // The data of each particle
	NumberOfParticles int32   // The number of particles to create
}

func (Particle) Id() (uint, uint) { return 0x2A, PktInvalid }

// 0x2B = Change Game State
type ChangeGameState struct {
	Reason uint8
	Value  float32 // Depends on reason
}

func (ChangeGameState) Id() (uint, uint) { return 0x2B, PktInvalid }

// 0x2C = Spawn Global Entity
type SpawnGlobalEntity struct {
	EntityID uint  // The entity ID of the thunderbolt
	Type     int8  // The global entity type, currently always 1 for thunderbolt.
	X        int32 // Thunderbolt X a [[Data_Types#Fixed-point_numbers|fixed-point number]]
	Y        int32 // Thunderbolt Y a [[Data_Types#Fixed-point_numbers|fixed-point number]]
	Z        int32 // Thunderbolt Z a [[Data_Types#Fixed-point_numbers|fixed-point number]]
}

func (SpawnGlobalEntity) Id() (uint, uint) { return 0x2C, PktInvalid }

// 0x2D = Open Window
type OpenWindow struct {
	WindowId               uint8  // A unique id number for the window to be displayed.  Notchian server implementation is a counter, starting at 1.
	InventoryType          uint8  // The window type to use for display.  Check below
	WindowTitle            string // The title of the window.
	NumberOfSlots          uint8  // Number of slots in the window (excluding the number of slots in the player inventory).
	UseProvidedWindowTitle bool   // If false, the client will look up a string like "window.minecart". If true, the client uses what the server provides.
	EntityID               int32  // EntityHorse's entityId. Only sent when window type is equal to 11 (AnimalChest).
}

func (OpenWindow) Id() (uint, uint) { return 0x2D, PktInvalid }

// 0x2E = Close Window (Clientbound)
// 0x0D = Close Window (Serverbound)
type CloseWindow struct {
	WindowID uint8 // This is the id of the window that was closed. 0 for inventory.
}

func (CloseWindow) Id() (uint, uint) { return 0x2E, 0x0D }

// 0x2F = Set Slot
type SetSlot struct {
	WindowID uint8 // The window which is being updated. 0 for player inventory. Note that all known window types include the player inventory. This packet will only be sent for the currently opened window while the player is performing actions, even if it affects the player inventory. After the window is closed, a number of these packets are sent to update the player's inventory window (0).
	Slot     int16 // The slot that should be updated
	SlotData Slot
}

func (SetSlot) Id() (uint, uint) { return 0x2F, PktInvalid }

// 0x30 = Window Items
type WindowItems struct {
	WindowID uint8  // The id of window which items are being sent for. 0 for player inventory.
	SlotData []Slot `mc:"len=int16"`
}

func (WindowItems) Id() (uint, uint) { return 0x30, PktInvalid }

// 0x31 = Window Property
type WindowProperty struct {
	WindowID uint8 // The id of the window.
	Property int16 // Which property should be updated.
	Value    int16 // The new value for the property.
}

func (WindowProperty) Id() (uint, uint) { return 0x31, PktInvalid }

// 0x32 = Confirm Transaction (Clientbound)
// 0x0F = Confirm Transaction (Serverbound)
type ConfirmTransaction struct {
	WindowID     uint8 // The id of the window that the action occurred in.
	ActionNumber int16 // Every action that is to be accepted has a unique number. This field corresponds to that number.
	Accepted     bool  // Whether the action was accepted.
}

func (ConfirmTransaction) Id() (uint, uint) { return 0x32, 0x0F }

// 0x33 = Update Sign (Clientbound)
// 0x12 = Update Sign (Serverbound)
type UpdateSign struct {
	X     int32  // Block X Coordinate
	Y     int16  // Block Y Coordinate
	Z     int32  // Block Z Coordinate
	Line1 string // First line of text in the sign
	Line2 string // Second line of text in the sign
	Line3 string // Third line of text in the sign
	Line4 string // Fourth line of text in the sign
}

func (UpdateSign) Id() (uint, uint) { return 0x33, 0x12 }

// 0x34 = Maps
type Maps struct {
	ItemDamage uint   // The damage value of the map being modified
	Data       []byte `mc:"len=int16"`
}

func (Maps) Id() (uint, uint) { return 0x34, PktInvalid }

// 0x35 = Update Block Entity
type UpdateBlockEntity struct {
	X       int32
	Y       int16
	Z       int32
	Action  uint8  // The type of update to perform
	NBTData []byte `mc:"len=int16"` // Present if data length > 0. Compressed with [[wikipedia:Gzip|gzip]]. Varies
}

func (UpdateBlockEntity) Id() (uint, uint) { return 0x35, PktInvalid }

// 0x36 = Sign Editor Open
type SignEditorOpen struct {
	X int32 // X in block coordinates
	Y int32 // Y in block coordinates
	Z int32 // Z in block coordinates
}

func (SignEditorOpen) Id() (uint, uint) { return 0x36, PktInvalid }

// 0x37 = Statistics
type Statistics struct {
	Values []StatisticsEntry
}

func (Statistics) Id() (uint, uint) { return 0x37, PktInvalid }

// 0x38 = Player List Item
type PlayerListItem struct {
	PlayerName string // Supports chat colouring, limited to 16 characters.
	Online     bool   // The client will remove the user from the list if false.
	Ping       int16  // Ping, presumably in ms.
}

func (PlayerListItem) Id() (uint, uint) { return 0x38, PktInvalid }

// 0x39 = Player Abilities (Clientbound)
// 0x13 = Player Abilities (Serverbound)
type PlayerAbilities struct {
	Flags        int8
	FlyingSpeed  float32 // previous integer value divided by 250
	WalkingSpeed float32 // previous integer value divided by 250
}

func (PlayerAbilities) Id() (uint, uint) { return 0x39, 0x13 }

// 0x3A = Tab-Complete
type TabCompleteResponse struct {
	Match []string // Possible Tab-Complete
}

func (TabCompleteResponse) Id() (uint, uint) { return 0x3A, PktInvalid }

// 0x3B = Scoreboard Objective
type ScoreboardObjective struct {
	ObjectiveName  string // An unique name for the objective
	ObjectiveValue string // The text to be displayed for the score.
	CreateRemove   int8   // 0 to create the scoreboard. 1 to remove the scoreboard. 2 to update the display text.
}

func (ScoreboardObjective) Id() (uint, uint) { return 0x3B, PktInvalid }

// 0x3C = Update Score
type UpdateScore struct {
	ItemName     string // An unique name to be displayed in the list.
	UpdateRemove int8   // 0 to create/update an item. 1 to remove an item.
	ScoreName    string // The unique name for the scoreboard to be updated. Only sent when Update/Remove does not equal 1.
	Value        int32  // The score to be displayed next to the entry. Only sent when Update/Remove does not equal 1.
}

func (UpdateScore) Id() (uint, uint) { return 0x3C, PktInvalid }

// 0x3D = Display Scoreboard
type DisplayScoreboard struct {
	Position  int8   // The position of the scoreboard. 0 = list, 1 = sidebar, 2 = belowName.
	ScoreName string // The unique name for the scoreboard to be displayed.
}

func (DisplayScoreboard) Id() (uint, uint) { return 0x3D, PktInvalid }

// 0x3E = Teams
type Teams struct {
	TeamName        string   // A unique name for the team. (Shared with scoreboard).
	Mode            int8     // If 0 then the team is created.
	TeamDisplayName string   // Only if Mode = 0 or 2.
	TeamPrefix      string   // Only if Mode = 0 or 2. Displayed before the players' name that are part of this team.
	TeamSuffix      string   // Only if Mode = 0 or 2. Displayed after the players' name that are part of this team.
	FriendlyFire    int8     // Only if Mode = 0 or 2; 0 for off, 1 for on, 3 for seeing friendly invisibles
	Players         []string `mc:"len=int16"` // Only if Mode = 0 or 3 or 4. Players to be added/remove from the team.
}

func (Teams) Id() (uint, uint) { return 0x3E, PktInvalid }

// 0x3F = Plugin Message
// 0x17 = Plugin Message
type PluginMessage struct {
	Channel string // Name of the "channel" used to send the data.
	Data    []byte `mc:"len=int16"` // Any data.
}

func (PluginMessage) Id() (uint, uint) { return 0x3F, 0x17 }

// 0x40 = Disconnect
type Disconnect struct {
	Reason string // Displayed to the client when the connection terminates. Must be valid JSON.
}

func (Disconnect) Id() (uint, uint) { return 0x40, PktInvalid }
