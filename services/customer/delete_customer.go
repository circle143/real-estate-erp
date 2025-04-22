package customer

//
//type hRemoveCustomerFromFlat struct{}
//
//func (rc *hRemoveCustomerFromFlat) validate(db *gorm.DB, orgId, society, flatId string) error {
//	societyInfoService := flat.CreateFlatSocietyInfoService(db, uuid.MustParse(flatId))
//	return common.IsSameSociety(societyInfoService, orgId, society)
//
//}
//func (rc *hRemoveCustomerFromFlat) execute(db *gorm.DB, orgId, society, flatId, customer string) error {
//	err := rc.validate(db, orgId, society, flatId)
//	if err != nil {
//		return err
//	}
//
//	return db.Transaction(func(tx *gorm.DB) error {
//		return nil
//	})
//}
//
//func (cs *customerService) removeCustomerFromFlat(w http.ResponseWriter, r *http.Request) {
//	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
//	societyRera := chi.URLParam(r, "society")
//	flatId := chi.URLParam(r, "flat")
//	customer := chi.URLParam(r, "customer")
//}