import (
	"path/filepath"

	utils "github.com/tempuslabs/s3s2_new/utils"
)

type File struct {
	Name     string
	Path     string
	Size     int64
	Modified time.Time
	Hash     string
}

// Method to customize which fields are represented in the JSON manifest
// We may want to store other attributes in the struct for the purpose of the script that
// we don't care about in the actual JSON manifest
func (f *File) MarshalJSON() ([]byte, error) {
	return json.Marshal(
	    &struct {Name string `json:"string"`}
	    {Name: f.Name}
	    )}

func GetCleanedName(f *File) string {
    return utils.ToSlashClean(f.Name)
}

func GetPosixName(f *File) string {
    return utils.ToPosixPath(f.GetCleanedName())
}

func GetRelativePath(f *File, relative_to string) string {
    return utils.GetRelativePath(f.Name, relative_to)
}

func GetFileStructsFromDir(directory string) []File {
	var file_structs []File

	err := filepath.Walk(options.Directory,
		func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			if !info.IsDir() && !strings.HasSuffix(path, "manifest.json") {

				log.Debugf("Instantiating file '%s' to manifest as name '%s'...", path, fmt_path)

				file_struct = File{
				Name: info.Name,
				Path: path,
				Size: info.Size(),
				Hash: hash(path, options)
				}

				file_structs = append(file_structs, file_struct)

	return files

}
