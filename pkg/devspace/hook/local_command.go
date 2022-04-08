package hook

import (
	"bytes"
	"context"
	"encoding/json"
	runtimevar "github.com/loft-sh/devspace/pkg/devspace/config/loader/variable/runtime"
	devspacecontext "github.com/loft-sh/devspace/pkg/devspace/context"
	"github.com/loft-sh/devspace/pkg/devspace/pipeline/engine"
	"io"
	"os"
	"strings"

	"github.com/loft-sh/devspace/pkg/devspace/config"
	"github.com/loft-sh/devspace/pkg/devspace/config/versions/latest"
	"github.com/loft-sh/devspace/pkg/devspace/dependency/types"
	"github.com/loft-sh/devspace/pkg/util/command"
	"github.com/pkg/errors"
)

func NewLocalCommandHook(stdout io.Writer, stderr io.Writer) Hook {
	return &localCommandHook{
		Stdout: stdout,
		Stderr: stderr,
	}
}

type localCommandHook struct {
	Stdout io.Writer
	Stderr io.Writer
}

func (l *localCommandHook) Execute(ctx devspacecontext.Context, hook *latest.HookConfig, cmdExtraEnv map[string]string) error {
	// Create extra env variables
	osArgsBytes, err := json.Marshal(os.Args)
	if err != nil {
		return err
	}
	extraEnv := map[string]string{
		OsArgsEnv: string(osArgsBytes),
	}
	if ctx.KubeClient() != nil {
		extraEnv[KubeContextEnv] = ctx.KubeClient().CurrentContext()
		extraEnv[KubeNamespaceEnv] = ctx.KubeClient().Namespace()
	}
	for k, v := range cmdExtraEnv {
		extraEnv[k] = v
	}

	// resolve hook command and args
	hookCommand, hookArgs, err := ResolveCommand(ctx.Context(), hook.Command, hook.Args, ctx.WorkingDir(), ctx.Config(), ctx.Dependencies())
	if err != nil {
		return err
	}

	// if args are nil we execute the command in a shell
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	defer func() {
		if hook.Name != "" {
			ctx.Config().SetRuntimeVariable("hooks."+hook.Name+".stdout", strings.TrimSpace(stdout.String()))
			ctx.Config().SetRuntimeVariable("hooks."+hook.Name+".stderr", strings.TrimSpace(stderr.String()))
		}
	}()

	if hook.Args == nil {
		return engine.ExecuteSimpleShellCommand(ctx.Context(), ctx.WorkingDir(), io.MultiWriter(l.Stdout, stdout), io.MultiWriter(l.Stderr, stderr), nil, extraEnv, hookCommand)
	}

	// else we execute it directly
	return command.CommandWithEnv(ctx.Context(), ctx.WorkingDir(), io.MultiWriter(l.Stdout, stdout), io.MultiWriter(l.Stderr, stderr), nil, extraEnv, hookCommand, hookArgs...)
}

func ResolveCommand(ctx context.Context, command string, args []string, dir string, config config.Config, dependencies []types.Dependency) (string, []string, error) {
	// resolve hook command
	hookCommand, err := runtimevar.NewRuntimeResolver(dir, true).FillRuntimeVariablesAsString(ctx, command, config, dependencies)
	if err != nil {
		return "", nil, errors.Wrap(err, "resolve image helpers")
	}

	// resolve args
	if args != nil {
		newArgs := []string{}
		for _, a := range args {
			newArg, err := runtimevar.NewRuntimeResolver(dir, true).FillRuntimeVariablesAsString(ctx, a, config, dependencies)
			if err != nil {
				return "", nil, errors.Wrap(err, "resolve image helpers")
			}

			newArgs = append(newArgs, newArg)
		}

		return hookCommand, newArgs, nil
	}

	return hookCommand, nil, nil
}
