package protocol

import (
	"reflect"
)

func init() {
	for i := 0; i < 4; i++ {
		for j := 0; j < 2; j++ {
			PacketData[i][j] = &HostState{
				HostType(j),
				CxnState(i),
				make(map[reflect.Type]PacketInfo),
				nil,
			}
		}
	}

	// serverpacket.go
	initPackets(StatePlay, Server,
		0x00, KeepAlive{},
		0x01, JoinGame{},
		0x02, ServerChatMessage{},
		0x03, TimeUpdate{},
		0x04, EntityEquipment{},
		0x05, SpawnPosition{},
		0x06, UpdateHealth{},
		0x07, Respawn{},
		0x08, ServerPlayerPositionAndLook{},
		0x09, ServerHeldItemChange{},
		0x0A, UseBed{},
		0x0B, ServerAnimation{},
		0x0C, SpawnPlayer{},
		0x0D, CollectItem{},
		0x0E, SpawnObject{},
		0x0F, SpawnMob{},
		0x10, SpawnPainting{},
		0x11, SpawnExperienceOrb{},
		0x12, EntityVelocity{},
		0x13, DestroyEntities{},
		0x14, Entity{},
		0x15, EntityRelativeMove{},
		0x16, EntityLook{},
		0x17, EntityLookAndRelativeMove{},
		0x18, EntityTeleport{},
		0x19, EntityHeadLook{},
		0x1A, EntityStatus{},
		0x1B, AttachEntity{},
		0x1C, EntityMetadata{},
		0x1D, EntityEffect{},
		0x1E, RemoveEntityEffect{},
		0x1F, SetExperience{},
		0x20, EntityProperties{},
		0x21, ChunkData{},
		0x22, MultiBlockChange{},
		0x23, BlockChange{},
		0x24, BlockAction{},
		0x25, BlockBreakAnimation{},
		0x26, MapChunkBulk{},
		0x27, Explosion{},
		0x28, Effect{},
		0x29, SoundEffect{},
		0x2A, Particle{},
		0x2B, ChangeGameState{},
		0x2C, SpawnGlobalEntity{},
		0x2D, OpenWindow{},
		0x2E, CloseWindow{},
		0x2F, SetSlot{},
		0x30, WindowItems{},
		0x31, WindowProperty{},
		0x32, ConfirmTransaction{},
		0x33, UpdateSign{},
		0x34, Maps{},
		0x35, UpdateBlockEntity{},
		0x36, SignEditorOpen{},
		0x37, Statistics{},
		0x38, PlayerListItem{},
		0x39, PlayerAbilities{},
		0x3A, TabCompleteResponse{},
		0x3B, ScoreboardObjective{},
		0x3C, UpdateScore{},
		0x3D, DisplayScoreboard{},
		0x3E, Teams{},
		0x3F, PluginMessage{},
		0x40, Disconnect{},
	)

	// clientpacket.go
	initPackets(StatePlay, Client,
		0x00, KeepAlive{},
		0x01, ClientChatMessage{},
		0x02, UseEntity{},
		0x03, Player{},
		0x04, PlayerPosition{},
		0x05, PlayerLook{},
		0x06, ClientPlayerPositionAndLook{},
		0x07, PlayerDigging{},
		0x08, PlayerBlockPlacement{},
		0x09, ClientHeldItemChange{},
		0x0A, ClientAnimation{},
		0x0B, EntityAction{},
		0x0C, SteerVehicle{},
		0x0D, CloseWindow{},
		0x0E, ClickWindow{},
		0x0F, ConfirmTransaction{},
		0x10, CreativeInventoryAction{},
		0x11, EnchantItem{},
		0x12, UpdateSign{},
		0x13, PlayerAbilities{},
		0x14, TabCompleteRequest{},
		0x15, ClientSettings{},
		0x16, ClientStatus{},
	)

	// handshakepacket.go
	initPackets(StateHandshake, Client,
		0x00, Handshake{},
	)

	initPackets(StateStatus, Server,
		0x00, StatusResponse{},
		0x01, StatusPing{},
	)

	initPackets(StateStatus, Client,
		0x00, StatusRequest{},
		0x01, StatusPing{},
	)

	initPackets(StateLogin, Server,
		0x00, LoginDisconnect{},
		0x01, EncryptionRequest{},
		0x02, LoginSuccess{},
	)

	initPackets(StateLogin, Client,
		0x00, LoginStart{},
		0x01, EncryptionResponse{},
	)
}

////////////////////////////////////////////////////////////////////////////////

type ReadFunc func(v reflect.Value, c *Coder)
type WriteFunc func(c *Coder, v reflect.Value)

// PacketInfo describes a packet
type PacketInfo struct {
	Id    int
	Rt    reflect.Type
	Read  ReadFunc
	Write WriteFunc
}

func initPackets(state CxnState, sender HostType, px ...interface{}) {
	for pi := 0; pi < len(px); pi += 2 {
		id, rt := px[pi].(int), reflect.TypeOf(px[pi+1])
		if rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		recver := 1 - sender
		vs := PacketData[int(state)][int(sender)]
		vr := PacketData[int(state)][int(recver)]
		if id <= len(vr.Recv) {
			vv := make([]*PacketInfo, id+len(vr.Recv)+1)
			copy(vv, vr.Recv)
			vr.Recv = vv
		}
		tc := cacheType(rt)
		pr, pw := tc.rf, tc.wf
		pinf := PacketInfo{Id: id, Rt: rt, Write: pw, Read: pr}
		vr.Recv[id] = &pinf
		vs.Send[rt] = pinf
	}
}
