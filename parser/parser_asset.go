package parser

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

func (record *FPakEntry) ReadUAsset(pak *PakFile, parser *PakParser) *FPackageFileSummary {
	// Skip UE4 pak header
	// TODO Find out what's in the pak header
	headerSize := int64(pak.Footer.HeaderSize())

	parser.Seek(headerSize+int64(record.FileOffset), 0)
	parser.Preload(int32(record.FileSize))

	tag := parser.ReadInt32()
	legacyFileVersion := parser.ReadInt32()
	legacyUE3Version := parser.ReadInt32()
	fileVersionUE4 := parser.ReadInt32()
	fileVersionLicenseeUE4 := parser.ReadInt32()

	// TODO custom_version_container: Vec<FCustomVersion>
	parser.Read(4)

	totalHeaderSize := parser.ReadInt32()
	folderName := parser.ReadString()
	packageFlags := parser.ReadUint32()
	nameCount := parser.ReadUint32()
	nameOffset := parser.ReadInt32()
	gatherableTextDataCount := parser.ReadInt32()
	gatherableTextDataOffset := parser.ReadInt32()
	exportCount := parser.ReadUint32()
	exportOffset := parser.ReadInt32()
	importCount := parser.ReadUint32()
	importOffset := parser.ReadInt32()
	dependsOffset := parser.ReadInt32()
	stringAssetReferencesCount := parser.ReadInt32()
	stringAssetReferencesOffset := parser.ReadInt32()
	searchableNamesOffset := parser.ReadInt32()
	thumbnailTableOffset := parser.ReadInt32()
	guid := parser.ReadFGuid()
	generationCount := parser.ReadUint32()

	generations := make([]*FGenerationInfo, generationCount)
	for i := uint32(0); i < generationCount; i++ {
		generations[i] = parser.ReadFGenerationInfo()
	}

	savedByEngineVersion := parser.ReadFEngineVersion()
	compatibleWithEngineVersion := parser.ReadFEngineVersion()
	compressionFlags := parser.ReadUint32()
	compressedChunkCount := parser.ReadUint32()

	compressedChunks := make([]*FCompressedChunk, compressedChunkCount)
	for i := uint32(0); i < compressedChunkCount; i++ {
		compressedChunks[i] = parser.ReadFCompressedChunk()
	}

	packageSource := parser.ReadUint32()
	additionalPackageCount := parser.ReadUint32()

	additionalPackagesToCook := make([]string, additionalPackageCount)
	for i := uint32(0); i < additionalPackageCount; i++ {
		additionalPackagesToCook[i] = parser.ReadString()
	}

	assetRegistryDataOffset := parser.ReadInt32()
	bulkDataStartOffset := parser.ReadInt32()
	worldTileInfoDataOffset := parser.ReadInt32()
	chunkCount := parser.ReadUint32()

	chunkIds := make([]int32, chunkCount)
	for i := uint32(0); i < chunkCount; i++ {
		chunkIds[i] = parser.ReadInt32()
	}

	// TODO unknown bytes
	parser.Read(4)

	preloadDependencyCount := parser.ReadInt32()
	preloadDependencyOffset := parser.ReadInt32()

	names := make([]*FNameEntrySerialized, nameCount)
	for i := uint32(0); i < nameCount; i++ {
		names[i] = &FNameEntrySerialized{
			Name:                  parser.ReadString(),
			NonCasePreservingHash: parser.ReadUint16(),
			CasePreservingHash:    parser.ReadUint16(),
		}
	}

	imports := make([]*FObjectImport, importCount)
	for i := uint32(0); i < importCount; i++ {
		imports[i] = &FObjectImport{
			ClassPackage: parser.ReadFName(names),
			ClassName:    parser.ReadFName(names),
			OuterIndex:   parser.ReadInt32(),
			ObjectName:   parser.ReadFName(names),
		}
	}

	exports := make([]*FObjectExport, exportCount)
	for i := uint32(0); i < exportCount; i++ {
		exports[i] = &FObjectExport{
			ClassIndex:                   parser.ReadFPackageIndex(imports, exports),
			SuperIndex:                   parser.ReadFPackageIndex(imports, exports),
			TemplateIndex:                parser.ReadFPackageIndex(imports, exports),
			OuterIndex:                   parser.ReadFPackageIndex(imports, exports),
			ObjectName:                   parser.ReadFName(names),
			Save:                         parser.ReadUint32(),
			SerialSize:                   parser.ReadInt64(),
			SerialOffset:                 parser.ReadInt64(),
			ForcedExport:                 parser.ReadInt32() != 0,
			NotForClient:                 parser.ReadInt32() != 0,
			NotForServer:                 parser.ReadInt32() != 0,
			PackageGuid:                  parser.ReadFGuid(),
			PackageFlags:                 parser.ReadUint32(),
			NotAlwaysLoadedForEditorGame: parser.ReadInt32() != 0,
			IsAsset:                      parser.ReadInt32() != 0,
			FirstExportDependency:        parser.ReadInt32(),
			SerializationBeforeSerializationDependencies: parser.ReadInt32() != 0,
			CreateBeforeSerializationDependencies:        parser.ReadInt32() != 0,
			SerializationBeforeCreateDependencies:        parser.ReadInt32() != 0,
			CreateBeforeCreateDependencies:               parser.ReadInt32() != 0,
		}
	}

	for _, objectImport := range imports {
		objectImport.OuterPackage = parser.ReadFPackageIndexInt(objectImport.OuterIndex, imports, exports)
	}

	// fmt.Println("UASSET LEFTOVERS:", len(fileData[offset:]))
	// fmt.Println(utils.HexDump(fileData[offset:]))

	// TODO Bunch of unknown bytes at the end

	return &FPackageFileSummary{
		Tag:                         tag,
		LegacyFileVersion:           legacyFileVersion,
		LegacyUE3Version:            legacyUE3Version,
		FileVersionUE4:              fileVersionUE4,
		FileVersionLicenseeUE4:      fileVersionLicenseeUE4,
		TotalHeaderSize:             totalHeaderSize,
		FolderName:                  folderName,
		PackageFlags:                packageFlags,
		NameOffset:                  nameOffset,
		GatherableTextDataCount:     gatherableTextDataCount,
		GatherableTextDataOffset:    gatherableTextDataOffset,
		ExportOffset:                exportOffset,
		ImportOffset:                importOffset,
		DependsOffset:               dependsOffset,
		StringAssetReferencesCount:  stringAssetReferencesCount,
		StringAssetReferencesOffset: stringAssetReferencesOffset,
		SearchableNamesOffset:       searchableNamesOffset,
		ThumbnailTableOffset:        thumbnailTableOffset,
		GUID:                        guid,
		Generations:                 generations,
		SavedByEngineVersion:        savedByEngineVersion,
		CompatibleWithEngineVersion: compatibleWithEngineVersion,
		CompressionFlags:            compressionFlags,
		CompressedChunks:            compressedChunks,
		PackageSource:               packageSource,
		AdditionalPackagesToCook:    additionalPackagesToCook,
		AssetRegistryDataOffset:     assetRegistryDataOffset,
		BulkDataStartOffset:         bulkDataStartOffset,
		WorldTileInfoDataOffset:     worldTileInfoDataOffset,
		ChunkIds:                    chunkIds,
		PreloadDependencyCount:      preloadDependencyCount,
		PreloadDependencyOffset:     preloadDependencyOffset,
		Names:                       names,
		Imports:                     imports,
		Exports:                     exports,
	}
}

func (record *FPakEntry) ReadUExp(pak *PakFile, parser *PakParser, uAsset *FPackageFileSummary) map[*FObjectExport]*ExportData {
	// Skip UE4 pak header
	// TODO Find out what's in the pak header
	headerSize := int64(pak.Footer.HeaderSize())

	exports := make(map[*FObjectExport]*ExportData)

	for _, export := range uAsset.Exports {
		offset := headerSize + int64(record.FileOffset) + (export.SerialOffset - int64(uAsset.TotalHeaderSize))
		log.Debugf("Reading export [%x]: %#v", offset, export.TemplateIndex.Reference)
		parser.Seek(offset, 0)

		tracker := parser.TrackRead()

		properties := parser.ReadFPropertyTagLoop(uAsset)

		if int64(tracker.bytesRead) < export.SerialSize {
			parser.Preload(int32(export.SerialSize - int64(tracker.bytesRead)))
		}

		parser.UnTrackRead()

		var data interface{}

		if parser.preload != nil {
			preloadSize := len(parser.preload)
			if preloadSize > 4 {
				var parsed bool
				data, parsed = parser.ReadClass(export, int32(preloadSize), uAsset)

				if !parsed {
					if className := export.TemplateIndex.ClassName(); className != nil {
						log.Warningf("Unknown export class type (%s)[%d]: %s", strings.Trim(export.ObjectName, "\x00"), preloadSize, strings.Trim(*className, "\x00"))
					}
				}
			}
		}

		exports[export] = &ExportData{
			Properties: properties,
			Data:       data,
		}
	}

	return exports
}

func (parser *PakParser) ReadFPropertyTag(uAsset *FPackageFileSummary, readData bool, depth int) *FPropertyTag {
	name := parser.ReadFName(uAsset.Names)

	if strings.Trim(name, "\x00") == "None" {
		return nil
	}

	propertyType := parser.ReadFName(uAsset.Names)
	size := parser.ReadInt32()
	arrayIndex := parser.ReadInt32()

	log.Tracef("%sReading Property %s (%s)[%d]", d(depth), strings.Trim(name, "\x00"), strings.Trim(propertyType, "\x00"), size)

	var tagData interface{}

	switch strings.Trim(propertyType, "\x00") {
	case "StructProperty":
		tagData = &StructProperty{
			Type: parser.ReadFName(uAsset.Names),
			Guid: parser.ReadFGuid(),
		}

		log.Tracef("%sStructProperty Type: %s", d(depth), tagData.(*StructProperty).Type)
		break
	case "BoolProperty":
		tagData = parser.Read(1)[0] != 0
		break
	case "EnumProperty":
		fallthrough
	case "ByteProperty":
		fallthrough
	case "SetProperty":
		fallthrough
	case "ArrayProperty":
		tagData = parser.ReadFName(uAsset.Names)
		break
	case "MapProperty":
		tagData = &MapProperty{
			KeyType:   parser.ReadFName(uAsset.Names),
			ValueType: parser.ReadFName(uAsset.Names),
		}
		break
	}

	hasGuid := parser.Read(1)[0] != 0

	var propertyGuid *FGuid

	if hasGuid {
		propertyGuid = parser.ReadFGuid()
	}

	var tag interface{}

	if readData && size > 0 {
		parser.Preload(size)
		tracker := parser.TrackRead()
		tag = parser.ReadTag(size, uAsset, propertyType, tagData, &name, depth)

		if tracker.bytesRead != size {
			log.Warningf("%sProperty not read correctly %s (%s)[%#v]: %d read out of %d",
				d(depth),
				strings.Trim(name, "\x00"),
				strings.Trim(propertyType, "\x00"),
				tagData,
				tracker.bytesRead,
				size)

			if tracker.bytesRead > size {
				log.Fatalf("More bytes read than available!")
			} else {
				parser.Read(size - tracker.bytesRead)
			}
		}

		parser.UnTrackRead()
	}

	return &FPropertyTag{
		Name:         name,
		PropertyType: propertyType,
		TagData:      tagData,
		Size:         size,
		ArrayIndex:   arrayIndex,
		PropertyGuid: propertyGuid,
		Tag:          tag,
	}
}

func (parser *PakParser) ReadTag(size int32, uAsset *FPackageFileSummary, propertyType string, tagData interface{}, name *string, depth int) interface{} {
	var tag interface{}
	switch strings.Trim(propertyType, "\x00") {
	case "FloatProperty":
		tag = parser.ReadFloat32()
		break
	case "ArrayProperty":
		arrayTypes := strings.Trim(tagData.(string), "\x00")
		valueCount := parser.ReadInt32()

		var innerTagData *FPropertyTag

		if arrayTypes == "StructProperty" {
			innerTagData = parser.ReadFPropertyTag(uAsset, false, depth+1)
		}

		values := make([]interface{}, valueCount)
		for i := int32(0); i < valueCount; i++ {
			switch arrayTypes {
			case "SoftObjectProperty":
				values[i] = &FSoftObjectPath{
					AssetPathName: parser.ReadFName(uAsset.Names),
					SubPath:       parser.ReadString(),
				}
				break
			case "StructProperty":
				log.Tracef("%sReading Array StructProperty: %s", d(depth), strings.Trim(innerTagData.TagData.(*StructProperty).Type, "\x00"))
				values[i] = &ArrayStructProperty{
					InnerTagData: innerTagData,
					Properties:   parser.ReadTag(-1, uAsset, arrayTypes, innerTagData.TagData, nil, depth+1),
				}
				break
			case "ObjectProperty":
				values[i] = parser.ReadFPackageIndex(uAsset.Imports, uAsset.Exports)
				break
			case "BoolProperty":
				values[i] = parser.Read(1)[0] != 0
				break
			case "ByteProperty":
				if (size-4)/valueCount == 1 {
					values[i] = parser.Read(1)[0]
				} else {
					values[i] = parser.ReadFName(uAsset.Names)
				}
				break
			case "NameProperty":
				fallthrough
			case "EnumProperty":
				values[i] = parser.ReadFName(uAsset.Names)
				break
			case "IntProperty":
				values[i] = parser.ReadInt32()
				break
			case "FloatProperty":
				values[i] = parser.ReadFloat32()
				break
			case "UInt32Property":
				values[i] = parser.ReadUint32()
				break
			case "TextProperty":
				values[i] = parser.ReadFText()
				break
			case "StrProperty":
				values[i] = parser.ReadString()
				break
			case "DelegateProperty":
				values[i] = &FScriptDelegate{
					Object: parser.ReadInt32(),
					Name:   parser.ReadFName(uAsset.Names),
				}
				break
			default:
				panic("unknown array type: " + arrayTypes)
			}
		}

		tag = values

		if valueCount > 0 && arrayTypes == "StructProperty" && values[0].(*ArrayStructProperty).Properties == nil {
			if size > 0 {
				// Struct data was not processed
				parser.Read(innerTagData.Size)
			}
		}

		break
	case "StructProperty":
		if tagData == nil {
			log.Trace("%sReading Generic StructProperty", d(depth))
		} else {
			log.Tracef("%sReading StructProperty: %s", d(depth), strings.Trim(tagData.(*StructProperty).Type, "\x00"))

			if structData, ok := tagData.(*StructProperty); ok {
				result, success := parser.ReadStruct(structData, size, uAsset, depth)

				if success {
					return &StructType{
						Type:  structData.Type,
						Value: result,
					}
				}
			}
		}

		properties := make([]*FPropertyTag, 0)

		for {
			property := parser.ReadFPropertyTag(uAsset, true, depth+1)

			if property == nil {
				break
			}

			properties = append(properties, property)
		}

		tag = properties
		break
	case "IntProperty":
		tag = parser.ReadInt32()
		break
	case "Int8Property":
		tag = int8(parser.Read(1)[0])
		break
	case "ObjectProperty":
		tag = parser.ReadFPackageIndex(uAsset.Imports, uAsset.Exports)
		break
	case "TextProperty":
		tag = parser.ReadFText()
		break
	case "BoolProperty":
		// No extra data
		break
	case "NameProperty":
		tag = parser.ReadFName(uAsset.Names)
		break
	case "StrProperty":
		tag = parser.ReadString()
		break
	case "UInt16Property":
		tag = parser.ReadUint16()
		break
	case "UInt32Property":
		tag = parser.ReadUint32()
		break
	case "UInt64Property":
		tag = parser.ReadUint64()
		break
	case "InterfaceProperty":
		tag = &UInterfaceProperty{
			InterfaceNumber: parser.ReadUint32(),
		}
		break
	case "ByteProperty":
		if size == 4 || size == -4 {
			tag = parser.ReadUint32()
		} else if size >= 8 {
			tag = parser.ReadFName(uAsset.Names)
		} else {
			tag = parser.Read(1)[0]
		}
		break
	case "SoftObjectProperty":
		tag = &FSoftObjectPath{
			AssetPathName: parser.ReadFName(uAsset.Names),
			SubPath:       parser.ReadString(),
		}
		break
	case "EnumProperty":
		if size == 8 {
			tag = parser.ReadFName(uAsset.Names)
		} else if size == 0 {
			break
		} else {
			panic("unknown state!")
		}
		break
	case "MapProperty":
		keyType := tagData.(*MapProperty).KeyType
		valueType := tagData.(*MapProperty).ValueType

		var keyData interface{}
		var valueData interface{}

		realTagData, ok := mapPropertyTypeOverrides[strings.Trim(*name, "\x00")]

		if ok {
			if strings.Trim(keyType, "\x00") != "StructProperty" {
				keyType = realTagData.KeyType
			} else {
				keyData = &StructProperty{
					Type: realTagData.KeyType,
				}
			}

			if strings.Trim(valueType, "\x00") != "StructProperty" {
				valueType = realTagData.ValueType
			} else {
				valueData = &StructProperty{
					Type: realTagData.ValueType,
				}
			}
		}

		if strings.Trim(keyType, "\x00") == "StructProperty" && keyData == nil {
			parser.Read(size)
			log.Warningf("%sSkipping MapProperty [%s] %s -> %s", d(depth), strings.Trim(*name, "\x00"), strings.Trim(keyType, "\x00"), strings.Trim(valueType, "\x00"))
			break
		}

		log.Tracef("%sReading MapProperty [%d]: %s -> %s", d(depth), size, strings.Trim(keyType, "\x00"), strings.Trim(valueType, "\x00"))

		numKeysToRemove := parser.ReadUint32()

		if numKeysToRemove != 0 {
			// TODO Read MapProperty where numKeysToRemove != 0
			parser.Read(size - 4)
			log.Warningf("%sSkipping MapProperty [%s] Remove Key Count: %d", d(depth), strings.Trim(*name, "\x00"), numKeysToRemove)
			break
		}

		num := parser.ReadInt32()

		results := make([]*MapPropertyEntry, num)
		for i := int32(0); i < num; i++ {
			key := parser.ReadTag(-4, uAsset, keyType, keyData, nil, depth+1)

			if key == nil {
				parser.Read(size - 8)
				log.Warningf("%sSkipping MapProperty [%s]: nil key", d(depth), strings.Trim(*name, "\x00"))
				break
			}

			value := parser.ReadTag(-4, uAsset, valueType, valueData, nil, depth+1)

			results[i] = &MapPropertyEntry{
				Key:   key,
				Value: value,
			}
		}

		tag = results
		break
	default:
		log.Debugf("%sUnread Tag Type: %s", d(depth), strings.Trim(propertyType, "\x00"))
		parser.Read(size)
		break
	}

	return tag
}
