package aws

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

func TestHumanizeString(t *testing.T) {
	if got, want := humanize(""), ""; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	if got, want := humanize("s"), "S"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	if got, want := humanize("STUFF"), "Stuff"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	if got, want := humanize("stuff"), "Stuff"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestDriver(t *testing.T) {
	awsMock := &mockEc2{}
	driv := NewDriver(awsMock)

	t.Run("Create vpc", func(t *testing.T) {
		cidr := "10.0.0.0/16"

		awsMock.verifyVpcInput = func(input *ec2.CreateVpcInput) error {
			if got, want := aws.StringValue(input.CidrBlock), cidr; got != want {
				return fmt.Errorf("got '%s', want '%s'", got, want)
			}
			return nil
		}

		id, err := driv.Create_Vpc(map[string]interface{}{"cidr": cidr})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := id.(string), "mynewvpc"; got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	})

	t.Run("Create subnet", func(t *testing.T) {
		cidr, vpc := "10.0.0.0/16", "anyvpc"

		awsMock.verifySubnetInput = func(input *ec2.CreateSubnetInput) error {
			if got, want := aws.StringValue(input.CidrBlock), cidr; got != want {
				return fmt.Errorf("got '%s', want '%s'", got, want)
			}
			if got, want := aws.StringValue(input.VpcId), vpc; got != want {
				return fmt.Errorf("got %s, want %s", got, want)
			}
			return nil
		}

		id, err := driv.Create_Subnet(map[string]interface{}{"cidr": cidr, "vpc": vpc})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := id.(string), "mynewsubnet"; got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	})

	t.Run("Create instance", func(t *testing.T) {
		countInt := 2
		image, typ, subnet, count, name := "ami-12", "t2.medium", "anysubnet", strconv.Itoa(countInt), "my_instance_name"
		var tagNameCreated bool

		awsMock.verifyInstanceInput = func(input *ec2.RunInstancesInput) error {
			if got, want := aws.StringValue(input.SubnetId), subnet; got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
			if got, want := aws.StringValue(input.ImageId), image; got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
			if got, want := aws.StringValue(input.InstanceType), typ; got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
			if got, want := aws.Int64Value(input.MinCount), int64(countInt); got != want {
				t.Fatalf("got %d, want %d", got, want)
			}
			if got, want := aws.Int64Value(input.MaxCount), int64(countInt); got != want {
				t.Fatalf("got %d, want %d", got, want)
			}
			return nil
		}

		awsMock.verifyTagInput = func(input *ec2.CreateTagsInput) error {
			if got, want := len(input.Tags), 1; got != want {
				t.Fatalf("got %d, want %d", got, want)
			}
			tagNameCreated = true
			if got, want := aws.StringValue(input.Tags[0].Key), "Name"; got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
			if got, want := aws.StringValue(input.Tags[0].Value), name; got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
			return nil
		}

		id, err := driv.Create_Instance(map[string]interface{}{"image": image, "type": typ, "subnet": subnet, "count": count, "name": name})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := id.(string), "mynewinstance"; got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
		if got, want := tagNameCreated, true; got != want {
			t.Fatalf("got %t, want %t", got, want)
		}
	})
}

type mockEc2 struct {
	ec2iface.EC2API
	verifyVpcInput      func(*ec2.CreateVpcInput) error
	verifySubnetInput   func(*ec2.CreateSubnetInput) error
	verifyInstanceInput func(*ec2.RunInstancesInput) error
	verifyTagInput      func(*ec2.CreateTagsInput) error
}

func (m *mockEc2) CreateVpc(input *ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	if err := m.verifyVpcInput(input); err != nil {
		return nil, err
	}
	return &ec2.CreateVpcOutput{Vpc: &ec2.Vpc{VpcId: aws.String("mynewvpc")}}, nil
}

func (m *mockEc2) CreateSubnet(input *ec2.CreateSubnetInput) (*ec2.CreateSubnetOutput, error) {
	if err := m.verifySubnetInput(input); err != nil {
		return nil, err
	}
	return &ec2.CreateSubnetOutput{Subnet: &ec2.Subnet{SubnetId: aws.String("mynewsubnet")}}, nil
}

func (m *mockEc2) RunInstances(input *ec2.RunInstancesInput) (*ec2.Reservation, error) {
	if err := m.verifyInstanceInput(input); err != nil {
		return nil, err
	}
	return &ec2.Reservation{Instances: []*ec2.Instance{&ec2.Instance{InstanceId: aws.String("mynewinstance")}}}, nil
}

func (m *mockEc2) CreateTags(input *ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error) {
	if err := m.verifyTagInput(input); err != nil {
		return nil, err
	}
	return &ec2.CreateTagsOutput{}, nil
}
