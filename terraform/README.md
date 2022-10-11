# Terraform config for Modron

Apply the terraform config for Modron:

```
cd dev # cd prod if you want to deploy to prod
tf plan -out tf.plan 
# Verify in the output that all the changes make sense and align with what you want to do.
tf apply "tf.plan"
```
