package modules

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const npm_registry = "https://registry.npmjs.org"

var tk_module_home = filepath.Join(".tk", "modules")

type npmManifest struct {
	Name     string                 `json:"name"`
	Tags     map[string]string      `json:"dist-tags"`
	Versions map[string]interface{} `json:"versions"`
}

// extractTarGzBytes extracts a tar.gz archive from a byte slice to a destination directory.
func extractTarGzBytes(tgzData []byte, destDir string) error {
	// Create a new bytes.Reader from the tgzData byte slice.
	byteReader := bytes.NewReader(tgzData)

	// Create a gzip.NewReader to decompress the gzip data.
	gzipReader, err := gzip.NewReader(byteReader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close() // Ensure the gzip reader is closed

	// Now that the data is decompressed, create a tar.NewReader.
	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Construct the full path for the extracted item.
		targetPath := filepath.Join(destDir, header.Name)

		// Handle different types of tar entries
		switch header.Typeflag {
		case tar.TypeDir:
			// If it's a directory, create it.
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			// If it's a regular file, create its parent directories if they don't exist.
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directories for file %s: %w", targetPath, err)
			}

			// Create the file.
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", targetPath, err)
			}
			defer outFile.Close() // Ensure the file is closed

			// Copy the content from the tar archive to the newly created file.
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("failed to copy file content to %s: %w", targetPath, err)
			}
		case tar.TypeSymlink:
			// Handle symbolic links
			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				return fmt.Errorf("failed to create symbolic link %s: %w", targetPath, err)
			}
		default:
			// Handle other types if necessary
			fmt.Printf("Skipping unknown tar entry type %c for %s\n", header.Typeflag, header.Name)
		}
	}
	return nil
}

// extractTarBytes extracts a tar archive from a byte slice to a destination directory.
func extractTarBytes(tarData []byte, destDir string) error {
	// Create a new bytes.Reader from the tarData byte slice.
	// This allows the tar.NewReader to read from the byte slice as if it were a file.
	tarReader := tar.NewReader(bytes.NewReader(tarData))

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Construct the full path for the extracted item.
		targetPath := filepath.Join(destDir, header.Name)

		// Handle different types of tar entries
		switch header.Typeflag {
		case tar.TypeDir:
			// If it's a directory, create it.
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			// If it's a regular file, create its parent directories if they don't exist.
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directories for file %s: %w", targetPath, err)
			}

			// Create the file.
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", targetPath, err)
			}
			defer outFile.Close() // Ensure the file is closed

			// Copy the content from the tar archive to the newly created file.
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("failed to copy file content to %s: %w", targetPath, err)
			}
		case tar.TypeSymlink:
			// Handle symbolic links
			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				return fmt.Errorf("failed to create symbolic link %s: %w", targetPath, err)
			}
		default:
			// Handle other types if necessary
			fmt.Printf("Skipping unknown tar entry type %c for %s\n", header.Typeflag, header.Name)
		}
	}
	return nil
}
func readManifest(specifier, version string) (*npmManifest, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s", npm_registry, specifier), nil)
	req.Header.Add("Content-Type", "application/json")
	res, e := http.DefaultClient.Do(req)
	if e != nil {
		return nil, e
	}
	defer res.Body.Close()
	data, e := io.ReadAll(res.Body)
	if e != nil {
		return nil, e
	}
	var manifest npmManifest
	e = json.Unmarshal(data, &manifest)
	if e != nil {
		return nil, e
	}
	return &manifest, nil
}

func download(uri, path string) error {
	res, e := http.DefaultClient.Get(uri)
	if e != nil {
		return e
	}
	defer res.Body.Close()
	data, e := io.ReadAll(res.Body)
	if e != nil {
		return e
	}
	home, _ := os.UserHomeDir()
	p := filepath.Join(home, path)
	e = os.MkdirAll(filepath.Dir(p), os.ModeDir)
	if e != nil {
		return e
	}
	return extractTarGzBytes(data, p)
}

func installFromNPM(specifier, version string) error {
	manifest, err := readManifest(specifier, version)
	if err != nil {
		return err
	}
	if m, ok := manifest.Versions[version]; ok {
		pkg := m.(nodePackage)
		if pkg.Dist != nil {
			return download(pkg.Dist.Tarball, fmt.Sprintf("./%s/%s", specifier, version))
		}
		return download(fmt.Sprintf("%s/-/%s-%s.tgz", npm_registry, specifier, version), filepath.Join(tk_module_home, specifier, version))
	} else {
		return fmt.Errorf("version %s not found for package %s", version, specifier)
	}
}
