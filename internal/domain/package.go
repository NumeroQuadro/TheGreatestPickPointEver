package domain

import "fmt"

type PackageType string

const (
	Box  PackageType = "box"
	Bag  PackageType = "bag"
	Film PackageType = "film"
)

const (
	BoxString  string = "box"
	BagString  string = "bag"
	FilmString string = "film"
)

func GetPackageTypeFromString(s string) (PackageType, error) {
	switch s {
	case BoxString:
		return Box, nil
	case BagString:
		return Bag, nil
	case FilmString:
		return Film, nil
	default:
		return PackageType(rune(0)), fmt.Errorf("unknown package type: %s", s)
	}
}

func GetStringFromPackageType(t PackageType) string {
	switch t {
	case Box:
		return BoxString
	case Bag:
		return BagString
	case Film:
		return FilmString
	default:
		return ""
	}
}

func IsPackageTypeValid(s string) bool {
	switch s {
	case BoxString:
		return true
	case BagString:
		return true
	case FilmString:
		return true
	default:
		return false
	}
}

type Package interface {
	ValidatePackagedOrder(weight int) error
	GetCost() int
}

type BoxPackage struct {
}

func (b *BoxPackage) GetCost() int {
	return 20
}

func (b *BoxPackage) ValidatePackagedOrder(weight int) error {
	const BoxPackageMaxWeight = 30

	if weight > BoxPackageMaxWeight {
		return ErrIncorrectWeightForApplyPackage
	}

	return nil
}

type FilmPackage struct {
}

func (f FilmPackage) ValidatePackagedOrder(_ int) error {
	return nil
}

func (f FilmPackage) GetCost() int {
	return 1
}

type BagPackage struct {
}

func (b *BagPackage) GetCost() int {
	return 5
}

func (b *BagPackage) ValidatePackagedOrder(weight int) error {
	const MaxWeightForBox = 10

	if weight > MaxWeightForBox {
		return ErrIncorrectWeightForApplyPackage
	}

	return nil
}

func GetPackagingStrategy(packageType PackageType) (Package, error) {
	var packageStrategy Package

	switch packageType {
	case Box:
		packageStrategy = &BoxPackage{}
	case Bag:
		packageStrategy = &BagPackage{}
	case Film:
		packageStrategy = &FilmPackage{}
	case "":
		break
	default:
		return nil, ErrPackageNotExists
	}

	return packageStrategy, nil
}
