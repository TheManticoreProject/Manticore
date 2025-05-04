package data_structures

import "github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_types"

// OBJECT_TYPE_LIST
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/6f04f1f2-d070-4f70-aae7-5f98ed63e1ba
type OBJECT_TYPE_LIST struct {
	// Level: Specifies the level of the object type in the hierarchy of an object and its sub-objects. Level zero
	// indicates the object itself. Level one indicates a sub-object of the object, such as a property set. Level two
	// indicates a sub-object of the level one sub-object, such as a property. There can be a maximum of five levels
	// numbered zero through four.
	Level data_types.WORD
	// Remaining: Remaining access bits for this element, used by the access check algorithm, as specified in section 2.5.3.2.
	Remaining data_types.UINT32 // data_types.ACCESS_MASK
	// ObjectType: A pointer to the GUID for the object or sub-object.
	ObjectType *GUID
}

type POBJECT_TYPE_LIST *OBJECT_TYPE_LIST
