//nolint:all
package firmware

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindExtractedRawLogo(t *testing.T) {
	t.Parallel()
	t.Run("context expired", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.TODO())
		cancel()
		cc := &ChecksumCollector{
			extractPath: "foo",
		}
		_, err := cc.findExtractedRawLogo(ctx)
		require.ErrorIs(t, err, context.Canceled)
	})
	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		cc := &ChecksumCollector{
			extractPath: t.TempDir(),
		}
		_, err := cc.findExtractedRawLogo(ctx)
		require.ErrorIs(t, err, errNoLogo)
	})
	t.Run("found it", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		rootDir := t.TempDir()
		err := os.MkdirAll(rootDir+"/foo/bar/baz", 0o750)
		require.NoError(t, err, "prerequisite dir setup 1")
		err = os.MkdirAll(rootDir+"/zip/zop/zoop/file-7bb28b99-61bb-11d5-9a5d-0090273fc14d/section0", 0o750)
		require.NoError(t, err, "prerequisite dir setup 2")
		logo, err := os.Create(rootDir + "/zip/zop/zoop/file-7bb28b99-61bb-11d5-9a5d-0090273fc14d/section0/section0.raw")
		require.NoError(t, err, "creating bogus logo")
		_, err = logo.WriteString("test logo file")
		require.NoError(t, err, "writing bogus logo")
		logo.Close()

		cc := &ChecksumCollector{
			extractPath: rootDir,
		}

		filename, err := cc.findExtractedRawLogo(ctx)
		require.NoError(t, err)
		require.Equal(t, "zip/zop/zoop/file-7bb28b99-61bb-11d5-9a5d-0090273fc14d/section0/section0.raw", filename)
	})
}

func TestHashDiscoveredLogo(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	err := os.MkdirAll(rootDir+"/zip/zop/zoop/file-7bb28b99-61bb-11d5-9a5d-0090273fc14d/section0", 0o750)
	require.NoError(t, err, "prerequisite dir setup")
	logo, err := os.Create(rootDir + "/zip/zop/zoop/file-7bb28b99-61bb-11d5-9a5d-0090273fc14d/section0/section0.raw")
	require.NoError(t, err, "creating bogus logo")
	_, err = logo.WriteString("test file data")
	require.NoError(t, err, "writing bogus logo")
	logo.Close()

	cc := &ChecksumCollector{
		extractPath: rootDir,
	}
	hash, err := cc.hashDiscoveredLogo(context.TODO(), "zip/zop/zoop/file-7bb28b99-61bb-11d5-9a5d-0090273fc14d/section0/section0.raw")
	require.NoError(t, err)
	require.Equal(t, "SHA256: 1be7aaf1938cc19af7d2fdeb48a11c381dff8a98d4c4b47b3b0a5044a5255c04", hash)
}
