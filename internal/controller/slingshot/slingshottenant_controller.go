/*
(C) Copyright Hewlett Packard Enterprise Development LP
*/

// Package slingshot defines the controller for slingshot tenant resource
package slingshot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.hpe.com/hpe/sshot-net-operator/httpclient"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	slingshot "github.hpe.com/hpe/sshot-net-operator/api/slingshot/v1alpha1"
	tapmsapi "github.hpe.com/hpe/sshot-net-operator/api/tapms/v1alpha2"
	tapms "github.hpe.com/hpe/sshot-net-operator/internal/controller/tapms"
	"github.hpe.com/hpe/sshot-net-operator/models"
)

// SlingshotTenantReconciler reconciles a SlingshotTenant object
type SlingshotTenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	httpClient                   = httpclient.NewClient(models.BaseURL)
	slingshotTenantGenerationMap = make(map[string]int64)
	slingshotTenantList          slingshot.SlingshotTenantList
	tenant                       tapmsapi.Tenant
)

//+kubebuilder:rbac:groups=slingshot.hpe.com.hpe.com,resources=slingshottenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=slingshot.hpe.com.hpe.com,resources=slingshottenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=slingshot.hpe.com.hpe.com,resources=slingshottenants/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SlingshotTenant object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *SlingshotTenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//get the slinghot tenant
	var sshotTenant slingshot.SlingshotTenant
	if err := r.Get(ctx, req.NamespacedName, &sshotTenant); err != nil {
		return ctrl.Result{}, nil
	}

	var tenants tapmsapi.TenantList
	if err := r.List(ctx, &tenants); err != nil {
		log.Printf("cannot get tenant: %+v", err)
		return ctrl.Result{}, err
	}

	//Get list of slingshot tenants
	var sTL slingshot.SlingshotTenantList
	if err := r.List(ctx, &sTL); err != nil {
		log.Printf("cannot list slingshot tenants: %s", err)
		return ctrl.Result{}, err
	}
	slingshotTenantList = sTL

	//Save tenant generation to map for update events
	for _, sTenant := range slingshotTenantList.Items {
		if _, ok := slingshotTenantGenerationMap[sTenant.Name]; !ok {
			slingshotTenantGenerationMap[sTenant.Name] = sTenant.Generation
		} else if sTenant.Generation < slingshotTenantGenerationMap[sTenant.Name] {
			slingshotTenantGenerationMap[sTenant.Name] = sTenant.Generation
		}
	}

	//find tenant with same name as slingshot tenant
	var tenantXnames []string
	var tenantFound bool
	for _, t := range tenants.Items {
		if t.Spec.TenantName == sshotTenant.Spec.TenantName {
			tenantFound = true
			tenant = t
			for _, tr := range t.Spec.TenantResources {
				tenantXnames = tr.XNames
				break
			}

		}
	}

	if !tenantFound {
		log.Printf("tenant not found: %s", sshotTenant.Spec.TenantName)
		return ctrl.Result{}, nil
	}

	if models.AccessToken == "" {
		log.Println("access token not found")
		return ctrl.Result{}, nil
	}

	//handle update event
	if sshotTenant.Generation != slingshotTenantGenerationMap[sshotTenant.Name] {
		err := r.handleUpdate(ctx, &sshotTenant, tenantXnames, httpClient)
		if err != nil {
			log.Printf("cannot update tenant: %s", err)
			return ctrl.Result{}, nil
		}

		slingshotTenantGenerationMap[sshotTenant.Name] = sshotTenant.Generation

	}

	return ctrl.Result{RequeueAfter: models.ReconciliationTime}, nil
}

var predicateFunctions = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		if e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration() {
			log.Printf("update event detected for %s/%s", e.ObjectNew.GetNamespace(), e.ObjectNew.GetName())
			return true
		}
		return false
	},
}

// SetupWithManager sets up the controller with the Manager.
func (r *SlingshotTenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&slingshot.SlingshotTenant{}).
		WithEventFilter(predicateFunctions).
		Complete(r)
}

func (r *SlingshotTenantReconciler) handleUpdate(ctx context.Context, instance *slingshot.SlingshotTenant, tenantXnames []string, httpClient *httpclient.Client) error {
	// to update VNI partition and VNI block, delete the VNI block and VNI partition and create them again

	log.Println("handling VNI update event for", instance.Spec.TenantName)

	//check if vni partition exists
	_, err := GetVNIPartition(instance.Spec.TenantName)
	if err != nil {
		log.Printf("cannot find VNI partition: %s", instance.Spec.TenantName)
		return err
	}

	var vniRequestData models.VNIRequestData
	vniRequestData.PartitionName = instance.Spec.TenantName
	vniRequestData.VNICount = instance.Spec.VNIPartition.VNICount
	vniRequestData.VNIRange = instance.Spec.VNIPartition.VNIRange

	// Validate the VNI request data
	err = tapms.ValidateVNIRequestData(vniRequestData)
	if err != nil {
		return err
	}

	vniBlockName := fmt.Sprintf("%s-%s", instance.Spec.TenantName, instance.Spec.VNIBlockName)

	err = tapms.HandleDelete(ctx, instance.Spec.TenantName, vniBlockName)
	if err != nil {
		log.Printf("cannot delete VNI partition and VNI block: %+v", err)
		return err
	}

	//recreate VNI partition and VNI block
	err = createVNIPartition(ctx, instance, tenantXnames, httpClient)
	if err != nil {
		log.Printf("cannot create VNI partition: %+v", err)
		return err
	}

	VNIBlock, err := tapms.CreateVNIBlock(ctx, tenant, *instance)
	if err != nil {
		log.Printf("cannot create VNI block: %+v", err)
		return err
	}

	log.Printf("updated VNI block %s for the tenant %s", VNIBlock.DocumentSelfLink, tenant.Spec.TenantName)

	//Check the stage of VniBlockEnforceTaskServiceState, keep checking until it is "FINISHED" or "FAILED"
	stage, err := tapms.CheckVniBlockEnforceTaskServiceState(ctx, VNIBlock.EnforcementTaskServiceLink)
	if err != nil {
		log.Printf("cannot check VniBlockEnforceTaskServiceState: %+v", err)
		return err
	}

	if stage {
		log.Printf("enforcement for VNI block %s is completed", VNIBlock.DocumentSelfLink)
	} else {
		log.Printf("enforcement for VNI block %s is failed", VNIBlock.DocumentSelfLink)
	}

	return nil

}

// GetVNIPartition gets the VNI partition
func GetVNIPartition(partition string) (models.VNIPartitionResponse, error) {
	ctx := context.Background()

	var vniPartition models.VNIPartitionResponse
	httpClient := httpclient.NewClient(models.BaseURL)
	responseBody, err := httpClient.SendRequest(ctx, "GET", "/fabric/vni/partitions/"+partition, models.VNIRequestData{})
	if err != nil {
		log.Printf("cannot get VNI partition: %+v %+v", partition, err)
		return vniPartition, err
	}

	err = json.Unmarshal(responseBody, &vniPartition)
	if err != nil {
		log.Printf("cannot unmarshal all VNI partitions: %+v %+v", responseBody, err)
		return vniPartition, err
	}

	return vniPartition, nil

}

// createVNIPartition creates the VNI partition
func createVNIPartition(ctx context.Context, instance *slingshot.SlingshotTenant, tenantXnames []string, httpClient *httpclient.Client) error {
	var vniRequestData models.VNIRequestData
	vniRequestData.PartitionName = instance.Spec.TenantName
	vniRequestData.VNICount = instance.Spec.VNIPartition.VNICount
	vniRequestData.VNIRange = instance.Spec.VNIPartition.VNIRange

	// Validate the VNI request data
	err := tapms.ValidateVNIRequestData(vniRequestData)
	if err != nil {
		return err
	}

	//Get edgePortDFAs for the tenant
	edgePortDFAList, _, err := tapms.GetEdgePortDFAList(tenantXnames)
	if err != nil {
		log.Printf("cannot get edge ports for tenant: %+v", err)
		return err
	}
	vniRequestData.EdgePortDFA = edgePortDFAList

	// Send the request
	_, err = httpClient.SendRequest(ctx, "POST", "/fabric/vni/partitions", vniRequestData)
	if err != nil {
		log.Printf("cannot create VNI partition: %+v", err)
		return err
	}

	log.Println("VNI partition created", vniRequestData.PartitionName)
	return nil
}
