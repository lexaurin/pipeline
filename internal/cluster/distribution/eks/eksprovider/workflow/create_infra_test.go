// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workflow

import (
	"testing"

	"emperror.dev/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"
)

type CreateInfraWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func TestCreateInfraWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(CreateInfraWorkflowTestSuite))
}

func (s *CreateInfraWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()

	createInfrastructureWorkflow := NewCreateInfrastructureWorkflow(nil)
	s.env.RegisterWorkflowWithOptions(createInfrastructureWorkflow.Execute, workflow.RegisterOptions{Name: CreateInfraWorkflowName})

	createVPCActivity := NewCreateVPCActivity(nil, "")
	s.env.RegisterActivityWithOptions(createVPCActivity.Execute, activity.RegisterOptions{Name: CreateVpcActivityName})

	createSubnetActivity := NewCreateSubnetActivity(nil, "")
	s.env.RegisterActivityWithOptions(createSubnetActivity.Execute, activity.RegisterOptions{Name: CreateSubnetActivityName})

	getSubnetsDetailsActivity := NewGetSubnetsDetailsActivity(nil)
	s.env.RegisterActivityWithOptions(getSubnetsDetailsActivity.Execute, activity.RegisterOptions{Name: GetSubnetsDetailsActivityName})

	createIamRolesActivity := NewCreateIamRolesActivity(nil, "")
	s.env.RegisterActivityWithOptions(createIamRolesActivity.Execute, activity.RegisterOptions{Name: CreateIamRolesActivityName})

	uploadSSHActivityActivity := NewUploadSSHKeyActivity(nil)
	s.env.RegisterActivityWithOptions(uploadSSHActivityActivity.Execute, activity.RegisterOptions{Name: UploadSSHKeyActivityName})

	createEksClusterActivity := NewCreateEksClusterActivity(nil)
	s.env.RegisterActivityWithOptions(createEksClusterActivity.Execute, activity.RegisterOptions{Name: CreateEksControlPlaneActivityName})

	saveK8sConfigActivity := NewSaveK8sConfigActivity(nil, nil)
	s.env.RegisterActivityWithOptions(saveK8sConfigActivity.Execute, activity.RegisterOptions{Name: SaveK8sConfigActivityName})

	getAMISizeActivity := NewGetAMISizeActivity(nil, nil)
	s.env.RegisterActivityWithOptions(getAMISizeActivity.Execute, activity.RegisterOptions{Name: GetAMISizeActivityName})

	selectVolumeSizeActivity := NewSelectVolumeSizeActivity(0)
	s.env.RegisterActivityWithOptions(selectVolumeSizeActivity.Execute, activity.RegisterOptions{Name: SelectVolumeSizeActivityName})

	createAsgActivity := NewCreateAsgActivity(nil, "", nil)
	s.env.RegisterActivityWithOptions(createAsgActivity.Execute, activity.RegisterOptions{Name: CreateAsgActivityName})

	createUserAccessKeyActivity := NewCreateClusterUserAccessKeyActivity(nil)
	s.env.RegisterActivityWithOptions(createUserAccessKeyActivity.Execute, activity.RegisterOptions{Name: CreateClusterUserAccessKeyActivityName})

	bootstrapActivity := NewBootstrapActivity(nil)
	s.env.RegisterActivityWithOptions(bootstrapActivity.Execute, activity.RegisterOptions{Name: BootstrapActivityName})

	saveClusterActivity := NewSaveNetworkDetailsActivity(nil)
	s.env.RegisterActivityWithOptions(saveClusterActivity.Execute, activity.RegisterOptions{Name: SaveNetworkDetailsActivityName})

	validateIAMRoleActivity := NewValidateIAMRoleActivity(nil)
	s.env.RegisterActivityWithOptions(validateIAMRoleActivity.Execute, activity.RegisterOptions{Name: ValidateIAMRoleActivityName})
}

func (s *CreateInfraWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *CreateInfraWorkflowTestSuite) Test_Successful_Create() {
	workflowInput := CreateInfrastructureWorkflowInput{
		Region:             "us-west-1",
		OrganizationID:     1,
		SecretID:           "my-secret-id",
		SSHSecretID:        "ssh-secret-id",
		ClusterID:          1,
		ClusterUID:         "cluster-id",
		ClusterName:        "test-cluster-name",
		VpcID:              "",
		RouteTableID:       "",
		VpcCidr:            "",
		ScaleEnabled:       false,
		DefaultUser:        false,
		ClusterRoleID:      "test-cluster-role-id",
		NodeInstanceRoleID: "test-node-instance-role-id",
		KubernetesVersion:  "1.14",
		EncryptionConfig: []EncryptionConfig{
			{
				Provider: Provider{
					KeyARN: "test-encryption-key-arn-or-alias",
				},
				Resources: []string{
					"test-resource-kind-1",
				},
			},
		},
		LogTypes:              []string{"test-log-type"},
		EndpointPublicAccess:  true,
		EndpointPrivateAccess: false,
		Subnets: []Subnet{
			{Cidr: "cidr1", AvailabilityZone: "az1"},
			{Cidr: "cidr2", AvailabilityZone: "az2"},
			{SubnetID: "subnet3"},
		},
		ASGSubnetMapping: map[string][]Subnet{
			"pool1": {
				{Cidr: "cidr1", AvailabilityZone: "az1"},
				{Cidr: "cidr2", AvailabilityZone: "az2"},
			},
			"pool2": {{SubnetID: "subnet3"}},
		},
		AsgList: []AutoscaleGroup{
			{
				Name:             "pool1",
				NodeSpotPrice:    "0.2",
				Autoscaling:      true,
				NodeMinCount:     2,
				NodeMaxCount:     3,
				Count:            2,
				NodeVolumeSize:   0,
				NodeImage:        "ami-test1",
				NodeInstanceType: "vm-type1-test",
				Labels: map[string]string{
					"test-label1":         "test-value1",
					"test-label2.io/name": "test-value2",
				},
			},
			{
				Name:             "pool2",
				NodeSpotPrice:    "0.0",
				Autoscaling:      false,
				NodeMinCount:     3,
				NodeMaxCount:     3,
				Count:            3,
				NodeVolumeSize:   12,
				NodeImage:        "ami-test2",
				NodeInstanceType: "vm-type2-test",
			},
		},
		UseGeneratedSSHKey: true,
	}

	eksActivity := EKSActivityInput{
		OrganizationID:            workflowInput.OrganizationID,
		SecretID:                  workflowInput.SecretID,
		Region:                    workflowInput.Region,
		ClusterName:               workflowInput.ClusterName,
		AWSClientRequestTokenBase: "default-test-workflow-id",
	}

	s.env.OnActivity(ValidateIAMRoleActivityName, mock.Anything, ValidateIAMRoleActivityInput{
		EKSActivityInput: eksActivity,
		ClusterRoleID:    workflowInput.ClusterRoleID,
	}).Return(&ValidateIAMRoleActivityOutput{}, nil)

	s.env.OnActivity(CreateIamRolesActivityName, mock.Anything, CreateIamRolesActivityInput{
		EKSActivityInput:   eksActivity,
		StackName:          "pipeline-eks-iam-test-cluster-name",
		DefaultUser:        workflowInput.DefaultUser,
		ClusterRoleID:      workflowInput.ClusterRoleID,
		NodeInstanceRoleID: workflowInput.NodeInstanceRoleID,
	},
	).Return(&CreateIamRolesActivityOutput{
		ClusterRoleArn:      "cluster-role-arn",
		ClusterUserArn:      "cluster-user-arn",
		NodeInstanceRoleID:  "node-instance-role-id",
		NodeInstanceRoleArn: "node-instance-role-arn",
	}, nil)

	s.env.OnActivity(CreateClusterUserAccessKeyActivityName, mock.Anything, CreateClusterUserAccessKeyActivityInput{
		EKSActivityInput: eksActivity,
		UserName:         "test-cluster-name",
		UseDefaultUser:   false,
		ClusterUID:       "cluster-id",
	}).Return(&CreateClusterUserAccessKeyActivityOutput{SecretID: "userSecretId"}, nil)

	s.env.OnActivity(UploadSSHKeyActivityName, mock.Anything, UploadSSHKeyActivityInput{
		EKSActivityInput: eksActivity,
		SSHKeyName:       "pipeline-eks-ssh-test-cluster-name",
		SSHSecretID:      "ssh-secret-id",
	}).Return(&UploadSSHKeyActivityOutput{}, nil)

	s.env.OnActivity(CreateVpcActivityName, mock.Anything, CreateVpcActivityInput{
		EKSActivityInput: eksActivity,
		StackName:        "pipeline-eks-test-cluster-name",
	}).Return(&CreateVpcActivityOutput{
		VpcID:               "new-vpc-id",
		RouteTableID:        "new-route-table-id",
		SecurityGroupID:     "test-eks-controlplane-security-group-id",
		NodeSecurityGroupID: "test-node-securitygroup-id",
	}, nil)

	s.env.OnActivity(CreateSubnetActivityName, mock.Anything, CreateSubnetActivityInput{
		EKSActivityInput: eksActivity,
		Cidr:             "cidr1",
		AvailabilityZone: "az1",
		StackName:        "pipeline-eks-subnet-test-cluster-name-cidr1",
		VpcID:            "new-vpc-id",
		RouteTableID:     "new-route-table-id",
	}).Return(&CreateSubnetActivityOutput{
		SubnetID:         "subnet1",
		Cidr:             "cidr1",
		AvailabilityZone: "az1",
	}, nil).Once()

	s.env.OnActivity(CreateSubnetActivityName, mock.Anything, CreateSubnetActivityInput{
		EKSActivityInput: eksActivity,
		Cidr:             "cidr2",
		AvailabilityZone: "az2",
		StackName:        "pipeline-eks-subnet-test-cluster-name-cidr2",
		VpcID:            "new-vpc-id",
		RouteTableID:     "new-route-table-id",
	}).Return(&CreateSubnetActivityOutput{
		SubnetID:         "subnet2",
		Cidr:             "cidr2",
		AvailabilityZone: "az2",
	}, nil).Once()

	s.env.OnActivity(GetSubnetsDetailsActivityName, mock.Anything, GetSubnetsDetailsActivityInput{
		OrganizationID: 1,
		SecretID:       "my-secret-id",
		Region:         "us-west-1",
		SubnetIDs: []string{
			"subnet3",
		},
	}).Return(&GetSubnetsDetailsActivityOutput{
		Subnets: []Subnet{
			{
				SubnetID:         "subnet3",
				Cidr:             "cidr3",
				AvailabilityZone: "az3",
			},
		},
	}, nil).Once()

	s.env.OnActivity(CreateEksControlPlaneActivityName, mock.Anything, CreateEksControlPlaneActivityInput{
		EKSActivityInput:  eksActivity,
		KubernetesVersion: "1.14",
		EncryptionConfig: []EncryptionConfig{
			{
				Provider: Provider{
					KeyARN: "test-encryption-key-arn-or-alias",
				},
				Resources: []string{
					"test-resource-kind-1",
				},
			},
		},
		EndpointPrivateAccess: false,
		EndpointPublicAccess:  true,
		ClusterRoleArn:        "cluster-role-arn",
		SecurityGroupID:       "test-eks-controlplane-security-group-id",
		LogTypes: []string{
			"test-log-type",
		},
		Subnets: []Subnet{
			{
				SubnetID:         "subnet1",
				Cidr:             "cidr1",
				AvailabilityZone: "az1",
			},
			{
				SubnetID:         "subnet2",
				Cidr:             "cidr2",
				AvailabilityZone: "az2",
			},
			{
				SubnetID:         "subnet3",
				Cidr:             "cidr3",
				AvailabilityZone: "az3",
			},
		},
	}).Return(&CreateEksControlPlaneActivityOutput{}, nil)

	s.env.OnActivity(GetAMISizeActivityName, mock.Anything, GetAMISizeActivityInput{
		EKSActivityInput: eksActivity,
		ImageID:          "ami-test1",
	}).Return(&GetAMISizeActivityOutput{AMISize: 4}, nil)

	s.env.OnActivity(SelectVolumeSizeActivityName, mock.Anything, SelectVolumeSizeActivityInput{
		AMISize:            4,
		OptionalVolumeSize: 0,
	}).Return(&SelectVolumeSizeActivityOutput{VolumeSize: 50}, nil)

	s.env.OnActivity(CreateAsgActivityName, mock.Anything, CreateAsgActivityInput{
		EKSActivityInput:    eksActivity,
		ClusterID:           1,
		StackName:           "pipeline-eks-nodepool-test-cluster-name-pool1",
		VpcID:               "new-vpc-id",
		SecurityGroupID:     "test-eks-controlplane-security-group-id",
		NodeSecurityGroupID: "test-node-securitygroup-id",
		NodeInstanceRoleID:  "node-instance-role-id",
		SSHKeyName:          "pipeline-eks-ssh-test-cluster-name",
		Name:                "pool1",
		NodeSpotPrice:       "0.2",
		Autoscaling:         true,
		NodeMinCount:        2,
		NodeMaxCount:        3,
		Count:               2,
		NodeVolumeSize:      50,
		NodeImage:           "ami-test1",
		NodeInstanceType:    "vm-type1-test",
		Labels: map[string]string{
			"test-label1":         "test-value1",
			"test-label2.io/name": "test-value2",
		},
		Subnets: []Subnet{
			{
				SubnetID:         "subnet1",
				Cidr:             "cidr1",
				AvailabilityZone: "az1",
			},
			{
				SubnetID:         "subnet2",
				Cidr:             "cidr2",
				AvailabilityZone: "az2",
			},
		},
	}).Return(&CreateAsgActivityOutput{}, nil).Once()

	s.env.OnActivity(GetAMISizeActivityName, mock.Anything, GetAMISizeActivityInput{
		EKSActivityInput: eksActivity,
		ImageID:          "ami-test2",
	}).Return(&GetAMISizeActivityOutput{AMISize: 8}, nil)

	s.env.OnActivity(SelectVolumeSizeActivityName, mock.Anything, SelectVolumeSizeActivityInput{
		AMISize:            8,
		OptionalVolumeSize: 12,
	}).Return(&SelectVolumeSizeActivityOutput{VolumeSize: 12}, nil)

	s.env.OnActivity(CreateAsgActivityName, mock.Anything, CreateAsgActivityInput{
		EKSActivityInput:    eksActivity,
		ClusterID:           1,
		StackName:           "pipeline-eks-nodepool-test-cluster-name-pool2",
		VpcID:               "new-vpc-id",
		SecurityGroupID:     "test-eks-controlplane-security-group-id",
		NodeSecurityGroupID: "test-node-securitygroup-id",
		NodeInstanceRoleID:  "node-instance-role-id",
		SSHKeyName:          "pipeline-eks-ssh-test-cluster-name",
		Name:                "pool2",
		NodeSpotPrice:       "0.0",
		Autoscaling:         false,
		NodeMinCount:        3,
		NodeMaxCount:        3,
		Count:               3,
		NodeVolumeSize:      12,
		NodeImage:           "ami-test2",
		NodeInstanceType:    "vm-type2-test",
		Subnets: []Subnet{
			{
				SubnetID:         "subnet3",
				Cidr:             "cidr3",
				AvailabilityZone: "az3",
			},
		},
	}).Return(&CreateAsgActivityOutput{}, nil).Once()

	s.env.OnActivity(BootstrapActivityName, mock.Anything, BootstrapActivityInput{
		EKSActivityInput:    eksActivity,
		KubernetesVersion:   "1.14",
		NodeInstanceRoleArn: "node-instance-role-arn",
		ClusterUserArn:      "cluster-user-arn",
	}).Return(&BootstrapActivityOutput{}, nil)

	s.env.OnActivity(SaveK8sConfigActivityName, mock.Anything, SaveK8sConfigActivityInput{
		ClusterID:        1,
		ClusterUID:       "cluster-id",
		ClusterName:      eksActivity.ClusterName,
		OrganizationID:   eksActivity.OrganizationID,
		Region:           eksActivity.Region,
		UserSecretID:     "userSecretId",
		ProviderSecretID: "my-secret-id",
	}).Return("userSecretId", nil)

	s.env.ExecuteWorkflow(CreateInfraWorkflowName, workflowInput)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *CreateInfraWorkflowTestSuite) Test_Successful_Fail_To_Create_VPC() {
	workflowInput := CreateInfrastructureWorkflowInput{
		Region:             "us-west-1",
		OrganizationID:     1,
		SecretID:           "my-secret-id",
		SSHSecretID:        "ssh-secret-id",
		ClusterUID:         "cluster-id",
		ClusterName:        "test-cluster-name",
		VpcID:              "",
		RouteTableID:       "",
		VpcCidr:            "",
		ScaleEnabled:       false,
		DefaultUser:        false,
		ClusterRoleID:      "test-cluster-role-id",
		NodeInstanceRoleID: "test-node-instance-role-id",
		KubernetesVersion:  "1.14",
		EncryptionConfig: []EncryptionConfig{
			{
				Provider: Provider{
					KeyARN: "test-encryption-key-arn-or-alias",
				},
				Resources: []string{
					"test-resource-kind-1",
				},
			},
		},
		LogTypes:              []string{"test-log-type"},
		EndpointPublicAccess:  true,
		EndpointPrivateAccess: false,
		Subnets: []Subnet{
			{Cidr: "cidr1", AvailabilityZone: "az1"},
			{Cidr: "cidr2", AvailabilityZone: "az2"},
			{SubnetID: "subnet3"},
		},
		ASGSubnetMapping: map[string][]Subnet{
			"pool1": {
				{Cidr: "cidr1", AvailabilityZone: "az1"},
				{Cidr: "cidr2", AvailabilityZone: "az2"},
			},
			"pool2": {{SubnetID: "subnet3"}},
		},
		AsgList: []AutoscaleGroup{
			{
				Name:             "pool1",
				NodeSpotPrice:    "0.2",
				Autoscaling:      true,
				NodeMinCount:     2,
				NodeMaxCount:     3,
				Count:            2,
				NodeVolumeSize:   0,
				NodeImage:        "ami-test1",
				NodeInstanceType: "vm-type1-test",
				Labels: map[string]string{
					"test-label1":         "test-value1",
					"test-label2.io/name": "test-value2",
				},
			},
			{
				Name:             "pool2",
				NodeSpotPrice:    "0.0",
				Autoscaling:      false,
				NodeMinCount:     3,
				NodeMaxCount:     3,
				Count:            3,
				NodeVolumeSize:   12,
				NodeImage:        "ami-test2",
				NodeInstanceType: "vm-type2-test",
			},
		},
		UseGeneratedSSHKey: true,
	}

	eksActivity := EKSActivityInput{
		OrganizationID:            workflowInput.OrganizationID,
		SecretID:                  workflowInput.SecretID,
		Region:                    workflowInput.Region,
		ClusterName:               workflowInput.ClusterName,
		AWSClientRequestTokenBase: "default-test-workflow-id",
	}

	s.env.OnActivity(ValidateIAMRoleActivityName, mock.Anything, ValidateIAMRoleActivityInput{
		EKSActivityInput: eksActivity,
		ClusterRoleID:    workflowInput.ClusterRoleID,
	}).Return(&ValidateIAMRoleActivityOutput{}, nil)

	s.env.OnActivity(CreateIamRolesActivityName, mock.Anything, mock.Anything).Return(&CreateIamRolesActivityOutput{
		ClusterRoleArn:      "cluster-role-arn",
		ClusterUserArn:      "cluster-user-arn",
		NodeInstanceRoleID:  "node-instance-role-id",
		NodeInstanceRoleArn: "node-instance-role-arn",
	}, nil)

	s.env.OnActivity(UploadSSHKeyActivityName, mock.Anything, mock.Anything).Return(&UploadSSHKeyActivityOutput{}, nil)

	s.env.OnActivity(CreateVpcActivityName, mock.Anything, mock.Anything).Return(nil, errors.New("failed to create VPC"))

	s.env.ExecuteWorkflow(CreateInfraWorkflowName, workflowInput)

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}
