package tfengine

import (
	"bytes"
	"context"       // Required by terraform-exec
	"encoding/json" // To potentially unmarshal output values
	"fmt"
	"log"
	"os"
	"os/exec" // Still needed for LookPath
	"path/filepath"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/pigen-dev/shared/utils"
)

type Terraform struct {
	Client *tfexec.Terraform
	varFilePath string
}

type TerraformFiles struct {
	MainTf []byte
	VariablesTf []byte
	OutputTf []byte
}
// setupTerraform creates a new Terraform executor instance.
func NewTF(tfVars map[string]any, TFFiles TerraformFiles, pluginLabel string) (*Terraform, error) {
	// Specify the path to the terraform executable
	execPath, err := exec.LookPath("terraform") // Attempts to find terraform in PATH
	if err != nil {
		return nil, fmt.Errorf("terraform executable not found in PATH: %w", err)
	}
	uniqueDir := fmt.Sprintf("./terraform/%s", pluginLabel)
	err = os.MkdirAll(uniqueDir, os.ModePerm)
	if err != nil {
			return nil, fmt.Errorf("Failed to create directory: %v", err)
	}
	varFilePath := filepath.Join(uniqueDir, "variables.tfvars.json")
	err= utils.TFVarParser(tfVars, varFilePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse variables: %v", err)
	}
	tf, err := tfexec.NewTerraform(uniqueDir, execPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create terraform executor: %w", err)
	}

	// Set stdout and stderr for the Terraform process for commands like init, plan, apply
	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)
	t := &Terraform{
		Client: tf,
		varFilePath: "variables.tfvars.json",
	}
	err = loadFiles(uniqueDir, TFFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to load Terraform files: %w", err)
	}
	return t, nil
}


func (t *Terraform)TerraformInit(ctx context.Context, projectID, pluginLabel string) error {
	backendBucket := fmt.Sprintf("bucket=%s-terraform-state-bucket", projectID)
	fmt.Println(backendBucket)
	backendPrefix := fmt.Sprintf("prefix=terraform/state/%s.tfstate", pluginLabel)
	log.Println("--------------------------------------------------")
	log.Println("Initializing Terraform...")
	log.Printf("Working Directory: %s\n", t.Client.WorkingDir())
	log.Println("--------------------------------------------------")
	err := t.Client.Init(
								ctx,
								tfexec.BackendConfig(backendBucket),
								tfexec.BackendConfig(backendPrefix),
							)

	if err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}
	log.Println("--- Terraform init finished successfully ---")
	log.Println("") // Add a newline for better separation
	return nil
}

// terraformPlan runs terraform plan using terraform-exec.
func (t *Terraform)TerraformPlan(ctx context.Context) error {
	log.Println("--------------------------------------------------")
	log.Println("Planning Terraform changes...")
	log.Printf("Working Directory: %s\n", t.Client.WorkingDir())
	log.Printf("Variable File: %s\n", t.varFilePath)
	log.Println("--------------------------------------------------")
	_, err := t.Client.Plan(ctx, tfexec.VarFile(t.varFilePath))
	if err != nil {
		return fmt.Errorf("terraform plan failed: %w", err)
	}

	log.Println("--- Terraform plan finished successfully ---")
	log.Println("") // Add a newline
	return nil
}

// terraformApply runs terraform apply using terraform-exec.
func (t *Terraform)TerraformApply(ctx context.Context) error {
	log.Println("--------------------------------------------------")
	log.Println("Applying Terraform changes...")
	log.Printf("Working Directory: %s\n", t.Client.WorkingDir())
	log.Printf("Variable File: %s\n", t.varFilePath)
	log.Println("--------------------------------------------------")

	// WARNING: Apply automatically approves.
	err := t.Client.Apply(ctx, tfexec.VarFile(t.varFilePath))
	if err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	log.Println("--- Terraform apply finished successfully ---")
	log.Println("") // Add a newline
	return nil
}

// terraformOutput retrieves and prints Terraform output values.
func (t *Terraform)TerraformOutput(ctx context.Context) (map[string] any,error) {
	log.Println("--------------------------------------------------")
	log.Println("Retrieving Terraform Outputs...")
	log.Printf("Working Directory: %s\n", t.Client.WorkingDir())
	log.Println("--------------------------------------------------")

	// Get the output values from the state
	outputs, err := t.Client.Output(ctx)
	if err != nil {
		// Common issue: Running output before apply if outputs depend on created resources.
		return nil, fmt.Errorf("terraform output failed: %w", err)
	}
	outputMap := make(map[string]any)
	if len(outputs) == 0 {
		log.Println("No outputs found in the Terraform state.")
	} else {
		log.Println("Outputs:")
		for key, outputMeta := range outputs {
			// outputMeta.Value is []byte containing the raw JSON value
			// outputMeta.Sensitive indicates if the output is marked as sensitive
			valueStr := string(outputMeta.Value)

			// Attempt to pretty-print if it's valid JSON, otherwise print as raw string
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, outputMeta.Value, "\t\t", "  "); err == nil {
				valueStr = prettyJSON.String()
			}

			if outputMeta.Sensitive {
				log.Printf("\t%s = <sensitive>\n", key)
			} else {
				log.Printf("\t%s = %s\n", key, valueStr)
			}
			outputMap[key] = string(outputMeta.Value)
		}
	}

	log.Println("--- Terraform output finished successfully ---")
	log.Println("") // Add a newline
	return outputMap, nil
}


func (t *Terraform)TerraformDestroy(ctx context.Context) error {
	log.Println("--------------------------------------------------")
	log.Println("Destroying Terraform Plugin...")
	log.Printf("Working Directory: %s\n", t.Client.WorkingDir())
	log.Printf("Variable File: %s\n", t.varFilePath)
	log.Println("--------------------------------------------------")

	// WARNING: Destroy automatically approves.
	err := t.Client.Destroy(ctx, tfexec.VarFile(t.varFilePath))
	if err != nil {
		return fmt.Errorf("terraform destroy failed: %w", err)
	}
	defer t.CleanUp()
	log.Println("Temp directory removed successfully.")
	log.Println("--- Terraform destroy finished successfully ---")
	log.Println("") // Add a newline
	return nil
}

func (t Terraform) CleanUp()error {
	// Remove the temp directory
	err := RemoveDir(t.Client.WorkingDir())
	if err != nil {
		return fmt.Errorf("Error removing temp directory: %v", err)
	}
	log.Println("Temp directory removed successfully.")
	return nil
}
func RemoveDir(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}
	return nil
}

func loadFiles(uniqueDir string, TFFiles TerraformFiles) error {
	mainPath := filepath.Join(uniqueDir, "main.tf")
	err := utils.WriteFile(mainPath, TFFiles.MainTf)
	if err != nil {
		return fmt.Errorf("Failed to write main.tf: %v", err)
	}
	varPath := filepath.Join(uniqueDir, "variables.tf")
	err = utils.WriteFile(varPath, TFFiles.VariablesTf)
	if err != nil {
		return fmt.Errorf("Failed to write variables.tf: %v", err)
	}
	outputPath := filepath.Join(uniqueDir, "output.tf")
	err = utils.WriteFile(outputPath, TFFiles.OutputTf)
	if err != nil {
		return fmt.Errorf("Failed to write output.tf: %v", err)
	}
	return nil
}