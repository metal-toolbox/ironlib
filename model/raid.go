package model

type CreateVirtualDiskOptions struct {
	RaidMode        string
	PhysicalDiskIDs []uint
	Name            string
	BlockSize       uint
}

type DestroyVirtualDiskOptions struct {
	VirtualDiskID int
}
