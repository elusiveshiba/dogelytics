{ pkgs ? import <nixpkgs> {} }:

let
  postgres = pkgs.postgresql_16;
  storageDirectory = "/storage";

  dogelytics-bin = pkgs.buildGoModule {
    pname = "dogelytics";
    version = "0.0.1";
    src = ./.;
    vendorHash = "sha256-gitjxJksapIE8tA7UxgfLLqudzk9Lq2sPy3mBFZF3QY=";

    nativeBuildInputs = [
      pkgs.go_1_24
    ];

    subPackages = [ "cmd/dogelytics" ];
  };

  dogelytics-run = pkgs.writeShellScriptBin "run.sh" ''
    set -eu

    PGDATA="${storageDirectory}/postgres"
    PGSOCKET="${storageDirectory}/postgres-run"
    PGPORT="5432"
    PGUSER="dogelytics"
    PGDATABASE="dogelytics"
    PGPASSFILE="${storageDirectory}/postgres-password.txt"
    SESSION_SECRET_FILE="${storageDirectory}/session-secret.txt"

    mkdir -p "$PGDATA" "$PGSOCKET"

    if [ ! -f "$PGPASSFILE" ]; then
      printf '%s\n' "''${DOGELYTICS_DB_PASSWORD:-dogebox_dogelytics_pup_temporary_static_password}" > "$PGPASSFILE"
      chmod 600 "$PGPASSFILE"
    fi
    PGPASSWORD="$(cat "$PGPASSFILE")"
    export PGPASSWORD

    if [ ! -f "$SESSION_SECRET_FILE" ]; then
      ${pkgs.openssl}/bin/openssl rand -hex 32 > "$SESSION_SECRET_FILE"
      chmod 600 "$SESSION_SECRET_FILE"
    fi

    if [ ! -s "$PGDATA/PG_VERSION" ]; then
      ${postgres}/bin/initdb \
        -D "$PGDATA" \
        --username=postgres \
        --auth-local=trust \
        --auth-host=scram-sha-256
    fi

    ${postgres}/bin/pg_ctl \
      -D "$PGDATA" \
      -o "-h 127.0.0.1 -k $PGSOCKET -p $PGPORT" \
      -w start

    stop_postgres() {
      ${postgres}/bin/pg_ctl -D "$PGDATA" -m fast -w stop >/dev/null 2>&1 || true
    }
    trap stop_postgres EXIT INT TERM

    if ! ${postgres}/bin/psql -h "$PGSOCKET" -p "$PGPORT" -U postgres -d postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname = '$PGUSER'" | ${pkgs.gnugrep}/bin/grep -q 1; then
      ${postgres}/bin/psql -h "$PGSOCKET" -p "$PGPORT" -U postgres -d postgres -c "CREATE USER $PGUSER WITH PASSWORD '$PGPASSWORD'"
    fi

    if ! ${postgres}/bin/psql -h "$PGSOCKET" -p "$PGPORT" -U postgres -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = '$PGDATABASE'" | ${pkgs.gnugrep}/bin/grep -q 1; then
      ${postgres}/bin/createdb -h "$PGSOCKET" -p "$PGPORT" -U postgres -O "$PGUSER" "$PGDATABASE"
    fi

    export BIND="''${DBX_PUP_IP}:4420"
    export INDEXER_API_URL="http://''${DBX_IFACE_INDEXER_API_HOST}:''${DBX_IFACE_INDEXER_API_PORT}"
    export DOGELYTICS_DBURL="postgres://$PGUSER:$PGPASSWORD@/$PGDATABASE?host=$PGSOCKET&sslmode=disable"
    export SESSION_SECRET="$(cat "$SESSION_SECRET_FILE")"
    export ADMIN_UI_PORT="4421"
    export ENABLE_DASHBOARD_UI="true"
    export DASHBOARD_UI_PORT="4422"

    dogelytics_pid=""
    stop_services() {
      if [ -n "$dogelytics_pid" ]; then
        kill "$dogelytics_pid" >/dev/null 2>&1 || true
        wait "$dogelytics_pid" >/dev/null 2>&1 || true
      fi
      stop_postgres
    }
    trap stop_services EXIT INT TERM

    ${dogelytics-bin}/bin/dogelytics &
    dogelytics_pid="$!"
    wait "$dogelytics_pid"
  '';

  dogelytics = pkgs.runCommand "dogelytics" {} ''
    mkdir -p "$out/bin"
    cp ${dogelytics-run}/bin/run.sh "$out/bin/run.sh"
    chmod +x "$out/bin/run.sh"
    cp -R ${./img} "$out/img"
  '';
in
{
  inherit dogelytics;
}
