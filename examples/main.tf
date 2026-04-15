terraform {
  required_providers {
    bicc = {
      source  = "tra-jkv/bicc"
      version = "~> 1.0"
    }
  }
}

# Configure the BICC Provider
provider "bicc" {
  host     = var.bicc_host     # Or use environment variable BICC_HOST
  username = var.bicc_username # Or use environment variable BICC_USERNAME
  password = var.bicc_password # Or use environment variable BICC_PASSWORD
}

# ============================================================================
# 1. BILLING ORDERS - Complete Sales Order Lifecycle
# Module: DooBiccExtractAM (Order Management)
# Use Case: OM is master for "Billing Orders" before they become AR Invoices
# Data Stores: Headers, Lines, Fulfillment, Addresses, Billing Plans, Document References, Payment
# ============================================================================

resource "bicc_job" "billing_orders" {
  name        = "BillingOrdersDev"
  description = "Complete billing order lifecycle - headers, lines, fulfillment, addresses, billing plans, document references, and payment"

  # Sales Order Headers
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.HeaderExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Sales Order Lines
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.LineExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Fulfillment Lines
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.FulfillLineExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Order Addresses
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.OrderAddressExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Billing Plans
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.BillingPlanExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Document References
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.DocumentReferencesExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Payment
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.PaymentExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
}

# ============================================================================
# 2. REPLENISHMENT ORDERS - Internal Transfer Orders
# Module: InvBiccExtractAM (Inventory Management)
# Use Case: Internal material transfers between warehouses/organizations
# Data Stores: Transfer Order Headers, Lines
# ============================================================================

resource "bicc_job" "replenishment_orders" {
  name        = "ReplenishmentOrdersDev"
  description = "Transfer orders for warehouse replenishment - headers and lines"

  # Transfer Order Headers
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.InvBiccExtractAM.TransferOrderHeaderExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Transfer Order Lines
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.InvBiccExtractAM.TransferOrderLineExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
}

# ============================================================================
# 3. RECEIVING - Advanced Shipment Notices (ASN)
# Module: RcvBiccExtractAM (Receiving)
# Use Case: ASN from suppliers, receipts, returns to vendor
# Data Stores: Inbound Shipment Headers, Lines
# ============================================================================

resource "bicc_job" "receiving" {
  name        = "ReceivingDev"
  description = "Receiving inbound shipments (ASN) - headers and lines"

  # Inbound Shipment Headers
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.RcvBiccExtractAM.ReceivingInboundShipmentHeaderExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Inbound Shipment Lines
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.RcvBiccExtractAM.ReceivingInboundShipmentLineExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
}

# ============================================================================
# 4. LOGISTICS - Carriers and Shipping
# Module: ScmRcsBiccExtractAM (Logistics)
# Use Case: Carrier master data for shipping and logistics
# Data Stores: Carrier
# ============================================================================

resource "bicc_job" "logistics" {
  name        = "LogisticsDev"
  description = "Carrier master data for shipping and logistics"

  # Carrier
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.ScmRcsBiccExtractAM.CarrierPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
}

# ============================================================================
# 5. PROCUREMENT ORDERS - Purchase Orders (Direct Spend)
# Module: PoBiccExtractAM, PrcPoPublicViewAM & FinFunBusinessUnitsAM (Procurement)
# Use Case: PO headers, lines, schedules for direct spend (with products)
# Data Stores: PO Headers, Lines, Distributions, ASL, Line Locations, Attribute Values, Procurement BU Usage, Style Headers
# ============================================================================

resource "bicc_job" "procurement_orders" {
  name        = "ProcurementOrdersDev"
  description = "Purchasing documents (POs, BPAs, etc.) - headers, lines, distributions, line locations, attribute values, style headers, and procurement BU usage"

  # Purchasing Document Headers (includes POs, BPAs, Contracts)
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingDocumentHeaderExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Purchasing Document Lines
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingDocumentLineExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Purchasing Document Distributions
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingDocumentDistributionExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Approved Supplier List (ASL)
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingASLExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Draft Purchase Order Distribution (for PO_LINE_LOCATIONS_ALL)
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcPoPublicViewAM.DraftPurchaseOrderDistributionPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Purchasing Attribute Values (for PO_ATTRIBUTE_VALUES)
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingAttributeValuesExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Procurement Business Unit Usage
  data_stores {
    data_store_key             = "FscmTopModelAM.FinFunBusinessUnitsAM.ProcurementBUUsagePVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Purchasing Agreement BU Access
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcPoPublicViewAM.PurchasingAgreementBuAccessPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Purchasing Document Line Location
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingDocumentLineLocationExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Purchasing Document Style Header
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingDocumentStyleHeaderExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
}

# ============================================================================
# 6. SUPPLIERS - Supplier Master Data
# Module: PozBiccExtractAM (Procurement)
# Use Case: Supplier master data, sites, contacts
# Data Stores: Suppliers, Supplier Sites
# ============================================================================

resource "bicc_job" "suppliers" {
  name        = "SuppliersDev"
  description = "Supplier master data - suppliers and sites"

  # Suppliers
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Supplier Sites
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierSiteExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # SupplierPVO
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcPozPublicViewAM.SupplierPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # SupplierSitePVO
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcPozPublicViewAM.SupplierSitePVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
}

# ============================================================================
# NOTE: Pricing and Customer modules are not available in this Oracle BICC 
# instance. These jobs have been removed from configuration.
# - Pricing: FscmTopModelAM.ScmExtractAM.QpBiccExtractAM not enabled
# - Customers: FscmTopModelAM.FndExtractAM.HzBiccExtractAM not enabled
# ============================================================================

# ============================================================================
# 7. PRODUCTS - Item/Product Master Data
# Module: EgpBiccExtractAM (Product Model)
# Use Case: Item master data - products, inventory items, attributes, categories, relationships
# Data Stores: Item Master, Item Category, Item Relationships
# ============================================================================

resource "bicc_job" "products" {
  name        = "ProductsDev"
  description = "Product master data - items, categories, and relationships"

  # Item Master
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.EgpBiccExtractAM.ItemExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Item Category
  data_stores {
    data_store_key             = "FscmTopModelAM.EgpItemsPublicModelAM.ItemCategory"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Item Relationships
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.EgpBiccExtractAM.ItemRelationshipExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Item Supplier Association
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.EgpBiccExtractAM.ItemSupplierAssociationExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
}

# ============================================================================
# 8. INVENTORY VALUATIONS - Cost Management
# Module: CstBiccExtractAM (Cost Management)
# Use Case: Financial valuation of inventory layers, costed inventory
# Data Stores: Inventory Valuations, Inventory On-hand
# ============================================================================

resource "bicc_job" "inventory_valuations" {
  name        = "InventoryValuationsDev"
  description = "Inventory on-hand valuations - financial valuation of inventory layers and on-hand quantities"

  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.CstBiccExtractAM.CstInventoryValuationExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Inventory On-hand
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.InvBiccExtractAM.InventoryOnhandExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
}

# ============================================================================
# 9. BPA - Blanket Purchase Agreements
# Module: PrcPoPublicViewAM (Procurement)
# Use Case: Blanket Purchase Agreements - long-term agreements with suppliers
# Data Stores: Agreement Headers, Agreement Lines
# ============================================================================

resource "bicc_job" "bpa" {
  name        = "BPADev"
  description = "Blanket Purchase Agreements - agreement headers and lines"

  # Agreement Headers
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcPoPublicViewAM.AgreementHeaderPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Agreement Lines
  # Filtered to include only columns from PO_LINE_LOCATIONS_ALL and PO_ATTRIBUTE_VALUES tables
  # Total: 114 columns (down from 1618)
  data_stores {
    data_store_key             = "FscmTopModelAM.PrcPoPublicViewAM.AgreementLinePVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true

    columns {
      name                = "FromBlanketDocumentTypeLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name                = "FromBlanketHeaderPrcBuLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name                = "FromBlanketHeaderStyleLineLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name                = "FromContractDocumentTypeLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name                = "FromContractHeaderPrcBuLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name                = "FromContractHeaderStyleLineLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name                = "FromHeaderLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name        = "FromLineLocationAccrueOnReceiptFlag"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationAllowSubstituteReceiptsFlag"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationAmount"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationAmountAccepted"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationAmountBilled"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationAmountCancelled"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationAmountFinanced"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationAmountReceived"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationAmountRecouped"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationAmountRejected"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationAmountShipped"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationAssessableValue"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationAutoClosureMode"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationBidPaymentId"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationCalculateTaxFlag"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationCancelDate"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationCancelFlag"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationCancelReason"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationCancelledBy"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationCarrierId"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationChangePromisedDateReason"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationClosedBy"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationClosedDate"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationClosedForInvoiceDate"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationClosedForReceivingDate"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationClosedReason"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationConsignedFlag"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationCountryOfOriginCode"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationDaysEarlyReceiptAllowed"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationDaysLateReceiptAllowed"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationDescription"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationDestinationTypeCode"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationDropShipFlag"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationEncumberNow"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationEncumberedDate"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationEncumberedFlag"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationEndDate"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationEnforceShipToLocationCode"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationEstimatedTaxAmount"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationFinalMatchFlag"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationFirmDate"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationFirmStatusLookupCode"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationFobLookupCode"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationFreightTermsLookupCode"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationFromHeaderId"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationFromLineId"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationFromLineLocationId"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationGovernmentContext"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationGroupName"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationInputTaxClassificationCode"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationInspectionRequiredFlag"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationInvoiceCloseTolerance"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationJobDefinitionName"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationJobDefinitionPackage"
      is_populate = true
    }

    columns {
      name        = "FromLineLocationLastAcceptDate"
      is_populate = true
    }

    columns {
      name                = "FromLineLocationLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name                = "HazardClassLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name                = "POSystemParametersLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name           = "PoLineId"
      is_populate    = true
      is_primary_key = true
    }

    columns {
      name        = "PurchasingAttributeValueAttachmentUrl"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueAttributeValuesId"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueAvailability"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueCreatedBy"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueCreationDate"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueJobDefinitionName"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueJobDefinitionPackage"
      is_populate = true
    }

    columns {
      name                = "PurchasingAttributeValueLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name        = "PurchasingAttributeValueLastUpdateLogin"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueLastUpdatedBy"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueLastUpdatedProgram"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueLeadTime"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueManufacturerPartNum"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueManufacturerUrl"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueObjectVersionNumber"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValuePicture"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValuePoHeaderId"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValuePoLineId"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValuePrcBuId"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueRebuildSearchIndexFlag"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueRequestId"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueSupplierUrl"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueThumbnailImage"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpAlias"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpAttributeValuesTlpId"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpComments"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpCreatedBy"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpCreationDate"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpDescription"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpJobDefinitionName"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpJobDefinitionPackage"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpLanguage"
      is_populate = true
    }

    columns {
      name                = "PurchasingAttributeValueTlpLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpLastUpdateLogin"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpLastUpdatedBy"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpLastUpdatedProgram"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpLongDescription"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpManufacturer"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpObjectVersionNumber"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpPoHeaderId"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpPoLineId"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpPrcBuId"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpRebuildSearchIndexFlag"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueTlpRequestId"
      is_populate = true
    }

    columns {
      name        = "PurchasingAttributeValueUnspsc"
      is_populate = true
    }

    columns {
      name                = "PurchasingDocumentHeaderLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name                = "PurchasingDocumentLineLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }

    columns {
      name                = "PurchasingDocumentVersionLastUpdateDate"
      is_populate         = true
      is_last_update_date = true
    }
  }
}

# ============================================================================
# 10. TRANSACTIONS - Receiving and Cost Transactions
# Module: CstBiccExtractAM & RcvBiccExtractAM (Cost Management & Receiving)
# Use Case: Cost transaction sources and receiving transaction details
# Data Stores: Incoming Transaction Cost Sources, Receipt Transactions
# ============================================================================

resource "bicc_job" "transactions" {
  name        = "TransactionsDev"
  description = "Cost and receiving transactions - incoming cost sources and receipt transactions"

  # Incoming Transaction Cost Sources
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.CstBiccExtractAM.CstIncomingTxnCostSourcesExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Receipt Transactions
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.RcvBiccExtractAM.ReceivingReceiptTransactionExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Inventory Transactions
  data_stores {
    data_store_key             = "FscmTopModelAM.ScmExtractAM.CstBiccExtractAM.CstInvTransactionsExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Receiving Receipt Transaction Reference
  data_stores {
    data_store_key             = "FscmTopModelAM.RcvReceiptsAM.ReceivingReceiptTransactionRefPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
}

# ============================================================================
# 11. HCM - Human Capital Management
# Module: HcmTopModelAnalyticsGlobalAM.HCMExtractAM (HCM)
# Use Case: Person and User data extraction
# Data Stores: Person Names, Users
# ============================================================================

resource "bicc_job" "hcm_person_user" {
  name        = "HCMPersonUserDev"
  description = "HCM person names and user data extraction"

  # Person Names
  data_stores {
    data_store_key             = "HcmTopModelAnalyticsGlobalAM.HCMExtractAM.PersonBiccExtractAM.PersonNameExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }

  # Users
  data_stores {
    data_store_key             = "HcmTopModelAnalyticsGlobalAM.HCMExtractAM.UserBiccExtractAM.UserExtractPVO"
    is_silent_error            = true
    is_effective_date_disabled = false
    use_union_for_incremental  = true
    auto_populate_all_columns  = true

    column_overrides {
      name                = "LastUpdateDate"
      is_last_update_date = true
    }
  }
}

# ============================================================================
# BACKFILLING EXAMPLES
# ============================================================================
# Use the enhanced backfill resource to backfill specific data stores within a job

# Example 1: Backfill all billing order data stores from June 2023
# resource "bicc_job_backfill" "billing_orders_backfill" {
#   job_id = bicc_job.billing_orders.id
#   
#   backfills {
#     data_store_key    = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.HeaderExtractPVO"
#     last_extract_date = "2023-06-01"
#   }
#   
#   backfills {
#     data_store_key    = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.LineExtractPVO"
#     last_extract_date = "2023-06-01"
#   }
#   
#   backfills {
#     data_store_key    = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.FulfillLineExtractPVO"
#     last_extract_date = "2023-06-01"
#   }
#   
#   backfills {
#     data_store_key    = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.OrderAddressExtractPVO"
#     last_extract_date = "2023-06-01"
#   }
#   
#   backfills {
#     data_store_key    = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.BillingPlanExtractPVO"
#     last_extract_date = "2023-06-01"
#   }
#   
#   backfills {
#     data_store_key    = "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.DocumentReferencesExtractPVO"
#     last_extract_date = "2023-06-01"
#   }
# }

# Example 2: Backfill only PO headers and lines (partial backfill)
# resource "bicc_job_backfill" "procurement_partial_backfill" {
#   job_id = bicc_job.procurement_orders.id
#   
#   backfills {
#     data_store_key    = "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchaseOrderHeaderExtractPVO"
#     last_extract_date = "2023-01-01"
#   }
#   
#   backfills {
#     data_store_key    = "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchaseOrderLineExtractPVO"
#     last_extract_date = "2023-01-01"
#   }
#   # Note: Shipments and ASL not backfilled - will extract all historical data
# }

# Example 3: Backfill with different dates per data store
# resource "bicc_job_backfill" "replenishment_staggered_backfill" {
#   job_id = bicc_job.replenishment_orders.id
#   
#   backfills {
#     data_store_key    = "FscmTopModelAM.ScmExtractAM.InvBiccExtractAM.TransferOrderHeaderExtractPVO"
#     last_extract_date = "2023-06-01"
#   }
#   
#   backfills {
#     data_store_key    = "FscmTopModelAM.ScmExtractAM.InvBiccExtractAM.TransferOrderLineExtractPVO"
#     last_extract_date = "2023-03-15"  # Different date for lines
#   }
# }

# ============================================================================
# OUTPUTS
# ============================================================================

output "job_ids" {
  description = "Map of job resource names to their IDs for reference in other resources"
  value = {
    billing_orders       = bicc_job.billing_orders.id
    replenishment_orders = bicc_job.replenishment_orders.id
    receiving            = bicc_job.receiving.id
    logistics            = bicc_job.logistics.id
    procurement_orders   = bicc_job.procurement_orders.id
    suppliers            = bicc_job.suppliers.id
    products             = bicc_job.products.id
    inventory_valuations = bicc_job.inventory_valuations.id
    bpa                  = bicc_job.bpa.id
    transactions         = bicc_job.transactions.id
  }
}

output "job_summary" {
  description = "Summary of all configured BICC jobs"
  value = {
    total_jobs        = 10
    total_data_stores = 33
    categories = {
      billing_orders = {
        job_name    = "BillingOrdersDev"
        data_stores = 7
        description = "Sales order headers, lines, fulfillment, addresses, billing plans, document references, payment"
      }
      replenishment_orders = {
        job_name    = "ReplenishmentOrdersDev"
        data_stores = 2
        description = "Transfer order headers and lines for warehouse replenishment"
      }
      receiving = {
        job_name    = "ReceivingDev"
        data_stores = 2
        description = "Inbound shipment headers and lines (ASN)"
      }
      logistics = {
        job_name    = "LogisticsDev"
        data_stores = 1
        description = "Carrier master data for shipping and logistics"
      }
      procurement_orders = {
        job_name    = "ProcurementOrdersDev"
        data_stores = 9
        description = "Purchasing documents (POs, BPAs, etc.), distributions, line locations, attribute values, style headers, approved supplier list, and procurement BU usage"
      }
      suppliers = {
        job_name    = "SuppliersDev"
        data_stores = 4
        description = "Supplier master data - suppliers and sites"
      }
      products = {
        job_name    = "ProductsDev"
        data_stores = 4
        description = "Product master data - items, categories, relationships"
      }
      inventory_valuations = {
        job_name    = "InventoryValuationsDev"
        data_stores = 2
        description = "Inventory on-hand valuations and on-hand quantities"
      }
      bpa = {
        job_name    = "BPADev"
        data_stores = 2
        description = "Blanket Purchase Agreements - agreement headers and lines"
      }
      transactions = {
        job_name    = "TransactionsDev"
        data_stores = 4
        description = "Cost and receiving transactions - incoming cost sources and receipt transactions"
      }
      hcm_person_user = {
        job_name    = "HCMPersonUserDev"
        data_stores = 2
        description = "HCM person names and user data extraction"
      }
    }
    note = "All jobs configured with incremental extraction using LastUpdateDate. Jobs grouped by functional category with multiple related data stores per job. No initial_extract_date set - will extract all historical data on first run. NOTE: Pricing and Customer modules not available in this Oracle instance."
  }
}

output "data_store_keys" {
  description = "Reference list of all data store keys for backfill configuration"
  value = {
    billing_orders = [
      "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.HeaderExtractPVO",
      "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.LineExtractPVO",
      "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.FulfillLineExtractPVO",
      "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.OrderAddressExtractPVO",
      "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.BillingPlanExtractPVO",
      "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.DocumentReferencesExtractPVO",
      "FscmTopModelAM.ScmExtractAM.DooBiccExtractAM.PaymentExtractPVO",
    ]
    replenishment_orders = [
      "FscmTopModelAM.ScmExtractAM.InvBiccExtractAM.TransferOrderHeaderExtractPVO",
      "FscmTopModelAM.ScmExtractAM.InvBiccExtractAM.TransferOrderLineExtractPVO",
    ]
    receiving = [
      "FscmTopModelAM.ScmExtractAM.RcvBiccExtractAM.ReceivingInboundShipmentHeaderExtractPVO",
      "FscmTopModelAM.ScmExtractAM.RcvBiccExtractAM.ReceivingInboundShipmentLineExtractPVO",
    ]
    logistics = [
      "FscmTopModelAM.ScmExtractAM.ScmRcsBiccExtractAM.CarrierPVO",
    ]
    procurement_orders = [
      "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingDocumentHeaderExtractPVO",
      "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingDocumentLineExtractPVO",
      "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingDocumentDistributionExtractPVO",
      "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingASLExtractPVO",
      "FscmTopModelAM.PrcPoPublicViewAM.DraftPurchaseOrderDistributionPVO",
      "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingAttributeValuesExtractPVO",
      "FscmTopModelAM.FinFunBusinessUnitsAM.ProcurementBUUsagePVO",
      "FscmTopModelAM.PrcPoPublicViewAM.PurchasingAgreementBuAccessPVO",
      "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingDocumentLineLocationExtractPVO",
      "FscmTopModelAM.PrcExtractAM.PoBiccExtractAM.PurchasingDocumentStyleHeaderExtractPVO",
    ]
    suppliers = [
      "FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierExtractPVO",
      "FscmTopModelAM.PrcExtractAM.PozBiccExtractAM.SupplierSiteExtractPVO",
      "FscmTopModelAM.PrcPozPublicViewAM.SupplierPVO",
      "FscmTopModelAM.PrcPozPublicViewAM.SupplierSitePVO"
    ]
    products = [
      "FscmTopModelAM.ScmExtractAM.EgpBiccExtractAM.ItemExtractPVO",
      "FscmTopModelAM.EgpItemsPublicModelAM.ItemCategory",
      "FscmTopModelAM.ScmExtractAM.EgpBiccExtractAM.ItemRelationshipExtractPVO",
      "FscmTopModelAM.ScmExtractAM.EgpBiccExtractAM.ItemSupplierAssociationExtractPVO",
    ]
    inventory_valuations = [
      "FscmTopModelAM.ScmExtractAM.CstBiccExtractAM.CstInventoryValuationExtractPVO",
      "FscmTopModelAM.ScmExtractAM.InvBiccExtractAM.InventoryOnhandExtractPVO",
    ]
    bpa = [
      "FscmTopModelAM.PrcPoPublicViewAM.AgreementHeaderPVO",
      "FscmTopModelAM.PrcPoPublicViewAM.AgreementLinePVO",
    ]
    transactions = [
      "FscmTopModelAM.ScmExtractAM.CstBiccExtractAM.CstIncomingTxnCostSourcesExtractPVO",
      "FscmTopModelAM.ScmExtractAM.RcvBiccExtractAM.ReceivingReceiptTransactionExtractPVO",
      "FscmTopModelAM.ScmExtractAM.CstBiccExtractAM.CstInvTransactionsExtractPVO",
      "FscmTopModelAM.RcvReceiptsAM.ReceivingReceiptTransactionRefPVO",
    ]
    hcm_person_user = [
      "HcmTopModelAnalyticsGlobalAM.HCMExtractAM.PersonBiccExtractAM.PersonNameExtractPVO",
      "HcmTopModelAnalyticsGlobalAM.HCMExtractAM.UserBiccExtractAM.UserExtractPVO",
    ]
  }
}
