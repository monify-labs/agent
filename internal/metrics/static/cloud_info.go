package static

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

// CloudInfo contains cloud provider metadata
type CloudInfo struct {
	Region       string
	InstanceType string
}

// DetectCloudProvider attempts to detect cloud provider and retrieve metadata
func DetectCloudProvider(ctx context.Context) (*CloudInfo, error) {
	// Try providers in order of popularity
	providers := []func(context.Context) (*CloudInfo, error){
		detectAWS,
		detectGCP,
		detectAzure,
		detectDigitalOcean,
	}

	for _, detector := range providers {
		if info, err := detector(ctx); err == nil && (info.Region != "" || info.InstanceType != "") {
			return info, nil
		}
	}

	// No cloud provider detected
	return &CloudInfo{}, nil
}

// detectAWS detects AWS EC2 instance metadata
func detectAWS(ctx context.Context) (*CloudInfo, error) {
	client := &http.Client{Timeout: 2 * time.Second}

	// AWS metadata endpoint
	baseURL := "http://169.254.169.254/latest/meta-data"

	// Get region
	region, err := fetchMetadata(ctx, client, baseURL+"/placement/availability-zone")
	if err != nil {
		return nil, err
	}
	// Remove availability zone suffix (e.g., us-east-1a -> us-east-1)
	if len(region) > 0 {
		region = region[:len(region)-1]
	}

	// Get instance type
	instanceType, _ := fetchMetadata(ctx, client, baseURL+"/instance-type")

	return &CloudInfo{
		Region:       region,
		InstanceType: instanceType,
	}, nil
}

// detectGCP detects Google Cloud Platform instance metadata
func detectGCP(ctx context.Context) (*CloudInfo, error) {
	client := &http.Client{Timeout: 2 * time.Second}

	// GCP requires this header
	baseURL := "http://metadata.google.internal/computeMetadata/v1/instance"

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/zone", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Zone format: projects/PROJECT_NUM/zones/ZONE
	zoneParts := strings.Split(string(body), "/")
	region := ""
	if len(zoneParts) > 0 {
		zone := zoneParts[len(zoneParts)-1]
		// Convert zone to region (e.g., us-central1-a -> us-central1)
		if idx := strings.LastIndex(zone, "-"); idx != -1 {
			region = zone[:idx]
		}
	}

	// Get machine type
	req, _ = http.NewRequestWithContext(ctx, "GET", baseURL+"/machine-type", nil)
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err = client.Do(req)
	instanceType := ""
	if err == nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		// Machine type format: projects/PROJECT_NUM/machineTypes/TYPE
		typeParts := strings.Split(string(body), "/")
		if len(typeParts) > 0 {
			instanceType = typeParts[len(typeParts)-1]
		}
	}

	return &CloudInfo{
		Region:       region,
		InstanceType: instanceType,
	}, nil
}

// detectAzure detects Azure instance metadata
func detectAzure(ctx context.Context) (*CloudInfo, error) {
	client := &http.Client{Timeout: 2 * time.Second}

	baseURL := "http://169.254.169.254/metadata/instance/compute"

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/location?api-version=2021-02-01&format=text", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Metadata", "true")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	region, _ := io.ReadAll(resp.Body)

	// Get VM size
	req, _ = http.NewRequestWithContext(ctx, "GET", baseURL+"/vmSize?api-version=2021-02-01&format=text", nil)
	req.Header.Set("Metadata", "true")
	resp, err = client.Do(req)
	instanceType := ""
	if err == nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		instanceType = string(body)
	}

	return &CloudInfo{
		Region:       string(region),
		InstanceType: instanceType,
	}, nil
}

// detectDigitalOcean detects DigitalOcean droplet metadata
func detectDigitalOcean(ctx context.Context) (*CloudInfo, error) {
	client := &http.Client{Timeout: 2 * time.Second}

	baseURL := "http://169.254.169.254/metadata/v1"

	region, err := fetchMetadata(ctx, client, baseURL+"/region")
	if err != nil {
		return nil, err
	}

	// DigitalOcean doesn't expose instance type via metadata
	return &CloudInfo{
		Region:       region,
		InstanceType: "",
	}, nil
}

// fetchMetadata is a helper to fetch metadata from a URL
func fetchMetadata(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}
