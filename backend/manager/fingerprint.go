package manager

import (
	"errors"
	"fmt"
	"io"

	"case/backend/manager/config"
	"case/backend/pkg/file"
	"case/backend/pkg/hash/md5"
	"case/backend/pkg/hash/oshash"
	"case/backend/pkg/logger"
	"case/backend/pkg/models"
)

type fingerprintCalculator struct {
	Config *config.Config
}

func (c *fingerprintCalculator) calculateOshash(f *models.BaseFile, o file.Opener) (*models.Fingerprint, error) {
	r, err := o.Open()
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	defer r.Close()

	rc, isRC := r.(io.ReadSeeker)
	if !isRC {
		return nil, errors.New("cannot calculate oshash for non-readcloser")
	}

	hash, err := oshash.FromReader(rc, f.Size)
	if err != nil {
		return nil, fmt.Errorf("calculating oshash: %w", err)
	}

	return &models.Fingerprint{
		Type:        models.FingerprintTypeOshash,
		Fingerprint: hash,
	}, nil
}

func (c *fingerprintCalculator) calculateMD5(o file.Opener) (*models.Fingerprint, error) {
	r, err := o.Open()
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	defer r.Close()

	hash, err := md5.FromReader(r)
	if err != nil {
		return nil, fmt.Errorf("calculating md5: %w", err)
	}

	return &models.Fingerprint{
		Type:        models.FingerprintTypeMD5,
		Fingerprint: hash,
	}, nil
}

func (c *fingerprintCalculator) CalculateFingerprints(f *models.BaseFile, o file.Opener, useExisting bool) ([]models.Fingerprint, error) {
	var ret []models.Fingerprint
	calculateMD5 := true

	if useAsVideo(f.Path) {
		var (
			fp  *models.Fingerprint
			err error
		)

		if useExisting {
			fp = f.Fingerprints.For(models.FingerprintTypeOshash)
		}

		if fp == nil {
			// calculate oshash first
			fp, err = c.calculateOshash(f, o)
			if err != nil {
				return nil, err
			}
		}

		ret = append(ret, *fp)

		// only calculate MD5 if enabled in config
		calculateMD5 = c.Config.IsCalculateMD5()
	}

	if calculateMD5 {
		var (
			fp  *models.Fingerprint
			err error
		)

		if useExisting {
			fp = f.Fingerprints.For(models.FingerprintTypeMD5)
		}

		if fp == nil {
			if useExisting {
				// log to indicate missing fingerprint is being calculated
				logger.Infof("Calculating checksum for %s ...", f.Path)
			}

			fp, err = c.calculateMD5(o)
			if err != nil {
				return nil, err
			}
		}

		ret = append(ret, *fp)
	}

	return ret, nil
}
