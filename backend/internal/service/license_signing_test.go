package service

import (
	"context"
	"testing"

	"license-manager/internal/models"
)

// productOfferingStub 是 ProductOfferingRepository 的最小桩，
// 记录 GetProductOfferingByID 调用次数以验证密钥查询收敛。
type productOfferingStub struct {
	offering *models.ProductOffering
	getErr   error
	getCount int
}

func (r *productOfferingStub) GetProductOfferingList(context.Context, *models.ProductOfferingListRequest) ([]models.ProductOffering, int64, error) {
	return nil, 0, nil
}
func (r *productOfferingStub) GetProductOfferingByID(context.Context, string) (*models.ProductOffering, error) {
	r.getCount++
	return r.offering, r.getErr
}
func (r *productOfferingStub) GetProductOfferingByCode(context.Context, string) (*models.ProductOffering, error) {
	return nil, nil
}
func (r *productOfferingStub) CreateProductOffering(context.Context, *models.ProductOffering) error {
	return nil
}
func (r *productOfferingStub) UpdateProductOffering(context.Context, *models.ProductOffering) error {
	return nil
}
func (r *productOfferingStub) UpdateSigningKeys(context.Context, *models.ProductOffering) error {
	return nil
}
func (r *productOfferingStub) DeleteProductOffering(context.Context, string) error {
	return nil
}
func (r *productOfferingStub) UpdateProductOfferingStatus(context.Context, string, string) error {
	return nil
}

// 收敛不变量：三类签名物料来自同一次 offering 查询。
func TestProductSigningMaterialSingleQuery(t *testing.T) {
	offeringID := "offering-1"
	offering := &models.ProductOffering{
		SigningPublicKey: "PUBKEY-CURRENT",
		SigningKeyChain:  models.JSON(`[]`),
	}
	stub := &productOfferingStub{offering: offering}
	svc := &licenseService{productRepo: stub}

	pub, keys, chain := svc.productSigningMaterial(&models.AuthorizationCode{ProductOfferingID: &offeringID})

	if pub != "PUBKEY-CURRENT" {
		t.Fatalf("public key = %q, want PUBKEY-CURRENT", pub)
	}
	if stub.getCount != 1 {
		t.Fatalf("GetProductOfferingByID called %d times, want 1（收敛后应只查一次）", stub.getCount)
	}
	_ = keys
	_ = chain
}

// 无产品项关联时三者应为空，且不触发查询。
func TestProductSigningMaterialNoProductOffering(t *testing.T) {
	stub := &productOfferingStub{}
	svc := &licenseService{productRepo: stub}

	pub, keys, chain := svc.productSigningMaterial(&models.AuthorizationCode{})

	if pub != "" || len(keys) > 0 || len(chain) > 0 {
		t.Fatalf("no product offering: pub=%q keys=%v chain=%v, want all empty", pub, keys, chain)
	}
	if stub.getCount != 0 {
		t.Fatalf("GetProductOfferingByID called %d times, want 0（无关联不应查询）", stub.getCount)
	}
}
