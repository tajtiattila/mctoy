package main

import (
	"reflect"
)

type Category uint

type CxnState byte

const (
	StateHandshake CxnState = 0
	StateStatus    CxnState = 1
	StateLogin     CxnState = 2
	StatePlay      CxnState = 3
)

type Direction byte

const (
	Clientbound Direction = 1
	Serverbound Direction = 2
)

type PktDisp struct {
	S CxnState
	D Direction
}

type Packet interface {
	Id(PktDisp) uint
}

////////////////////////////////////////////////////////////////////////////////

type stateInfo struct {
	ClientToServer []packetInfo
	ServerToClient []packetInfo
}

type packetInfo struct {
	Id   uint
	Name string
	New  pktfn
}

type pktfn func() Packet

var packets []*stateInfo

////////////////////////////////////////////////////////////////////////////////

type stateInit struct {
	ClientToServer []packetInit
	ServerToClient []packetInit
}

type packetInit struct {
	New pktfn
}

var packetinit = []stateInit{
	// StateHandshake
	{
		ServerToClient: nil,
		ClientToServer: []packetInit{
			{func() Packet { return new(Handshake) }},
		},
	},
	// StateStatus
	{
		ServerToClient: []packetInit{
			{func() Packet { return new(StatusResponse) }},
			{func() Packet { return new(StatusPing) }},
		},
		ClientToServer: []packetInit{
			{func() Packet { return new(StatusRequest) }},
			{func() Packet { return new(StatusPing) }},
		},
	},
	// StateLogin
	{
		ServerToClient: []packetInit{
			{func() Packet { return new(Disconnect) }},
			{func() Packet { return new(EncryptionRequest) }},
			{func() Packet { return new(LoginSuccess) }},
		},
		ClientToServer: []packetInit{
			{func() Packet { return new(LoginStart) }},
			{func() Packet { return new(EncryptionResponse) }},
		},
	},
	// StatePlay
	{
		ServerToClient: []packetInit{
			{func() Packet { return new(KeepAlive) }},
			{func() Packet { return new(JoinGame) }},
			{func() Packet { return new(ServerChatMessage) }},
			{func() Packet { return new(TimeUpdate) }},
			{func() Packet { return new(EntityEquipment) }},
			{func() Packet { return new(SpawnPosition) }},
			{func() Packet { return new(UpdateHealth) }},
			{func() Packet { return new(Respawn) }},
			{func() Packet { return new(ServerPlayerPositionAndLook) }},
			{func() Packet { return new(ServerHeldItemChange) }},
			{func() Packet { return new(UseBed) }},
			{func() Packet { return new(ServerAnimation) }},
			{func() Packet { return new(SpawnPlayer) }},
			{func() Packet { return new(CollectItem) }},
			{func() Packet { return new(SpawnObject) }},
			{func() Packet { return new(SpawnMob) }},
			{func() Packet { return new(SpawnPainting) }},
			{func() Packet { return new(SpawnExperienceOrb) }},
			{func() Packet { return new(EntityVelocity) }},
			{func() Packet { return new(DestroyEntities) }},
			{func() Packet { return new(Entity) }},
			{func() Packet { return new(EntityRelativeMove) }},
			{func() Packet { return new(EntityLook) }},
			{func() Packet { return new(EntityLookAndRelativeMove) }},
			{func() Packet { return new(EntityTeleport) }},
			{func() Packet { return new(EntityHeadLook) }},
			{func() Packet { return new(EntityStatus) }},
			{func() Packet { return new(AttachEntity) }},
			{func() Packet { return new(EntityMetadata) }},
			{func() Packet { return new(EntityEffect) }},
			{func() Packet { return new(RemoveEntityEffect) }},
			{func() Packet { return new(SetExperience) }},
			{func() Packet { return new(EntityProperties) }},
			{func() Packet { return new(ChunkData) }},
			{func() Packet { return new(MultiBlockChange) }},
			{func() Packet { return new(BlockChange) }},
			{func() Packet { return new(BlockAction) }},
			{func() Packet { return new(BlockBreakAnimation) }},
			{func() Packet { return new(MapChunkBulk) }},
			{func() Packet { return new(Explosion) }},
			{func() Packet { return new(Effect) }},
			{func() Packet { return new(SoundEffect) }},
			{func() Packet { return new(Particle) }},
			{func() Packet { return new(ChangeGameState) }},
			{func() Packet { return new(SpawnGlobalEntity) }},
			{func() Packet { return new(OpenWindow) }},
			{func() Packet { return new(CloseWindow) }},
			{func() Packet { return new(SetSlot) }},
			{func() Packet { return new(WindowItems) }},
			{func() Packet { return new(WindowProperty) }},
			{func() Packet { return new(ConfirmTransaction) }},
			{func() Packet { return new(UpdateSign) }},
			{func() Packet { return new(Maps) }},
			{func() Packet { return new(UpdateBlockEntity) }},
			{func() Packet { return new(SignEditorOpen) }},
			{func() Packet { return new(Statistics) }},
			{func() Packet { return new(PlayerListItem) }},
			{func() Packet { return new(PlayerAbilities) }},
			{func() Packet { return new(TabCompleteResponse) }},
			{func() Packet { return new(ScoreboardObjective) }},
			{func() Packet { return new(UpdateScore) }},
			{func() Packet { return new(DisplayScoreboard) }},
			{func() Packet { return new(Teams) }},
			{func() Packet { return new(PluginMessage) }},
			{func() Packet { return new(Disconnect) }},
		},
		ClientToServer: []packetInit{
			{func() Packet { return new(KeepAlive) }},
			{func() Packet { return new(ClientChatMessage) }},
			{func() Packet { return new(UseEntity) }},
			{func() Packet { return new(Player) }},
			{func() Packet { return new(PlayerPosition) }},
			{func() Packet { return new(PlayerLook) }},
			{func() Packet { return new(ClientPlayerPositionAndLook) }},
			{func() Packet { return new(PlayerDigging) }},
			{func() Packet { return new(PlayerBlockPlacement) }},
			{func() Packet { return new(ClientHeldItemChange) }},
			{func() Packet { return new(ClientAnimation) }},
			{func() Packet { return new(EntityAction) }},
			{func() Packet { return new(SteerVehicle) }},
			{func() Packet { return new(CloseWindow) }},
			{func() Packet { return new(ClickWindow) }},
			{func() Packet { return new(ConfirmTransaction) }},
			{func() Packet { return new(CreativeInventoryAction) }},
			{func() Packet { return new(EnchantItem) }},
			{func() Packet { return new(UpdateSign) }},
			{func() Packet { return new(PlayerAbilities) }},
			{func() Packet { return new(TabCompleteRequest) }},
			{func() Packet { return new(ClientSettings) }},
			{func() Packet { return new(ClientStatus) }},
			{func() Packet { return new(PluginMessage) }},
		},
	},
}

func addPacketInfo(d PktDisp, pfn pktfn) {
	v := packets[d.S]
	if v == nil {
		v = new(stateInfo)
		packets[d.S] = v
	}
	var vv []packetInfo
	if d.D == Serverbound {
		vv = v.ClientToServer
	} else {
		vv = v.ServerToClient
	}
	packet := pfn()
	id := packet.Id(d)
	name := reflect.ValueOf(packet).Elem().Type().Name()
	if cap(vv) <= int(id) {
		n := int(id) + 1
		vvn := make([]packetInfo, n*3/2+1)
		copy(vvn, vv)
		vv = vvn
	}
	vv[int(id)] = packetInfo{id, name, pfn}
	if d.D == Serverbound {
		v.ClientToServer = vv
	} else {
		v.ServerToClient = vv
	}
}

func init() {
	packets = make([]*stateInfo, 4)
	for s, si := range packetinit {
		for _, pi := range si.ServerToClient {
			addPacketInfo(PktDisp{S: CxnState(s), D: Clientbound}, pi.New)
		}
		for _, pi := range si.ClientToServer {
			addPacketInfo(PktDisp{S: CxnState(s), D: Serverbound}, pi.New)
		}
	}
}
